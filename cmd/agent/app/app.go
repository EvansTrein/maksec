//go:build linux

// Агент наблюдения за скриптами на целевом Linux-хосте.
// Механизм отслеживания — fanotify: ядро само присылает события об открытии
// (FAN_OPEN) и запуске (FAN_OPEN_EXEC) файлов в наблюдаемом каталоге.
// Сборка только под Linux (см. Makefile, target build-agent).
package app

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"strings"
	"time"

	"maksec/internal/config"
	"maksec/internal/entity"
	"maksec/pkg/logger"

	"github.com/rs/zerolog"
	"golang.org/x/sys/unix"
)

const (
	callbackTimeout  = 5 * time.Second
	fanotifyMetaSize = 24
)

type Option func(a *Agent)

type Agent struct {
	log              *zerolog.Logger
	callbackEndpoint string
	watchDir         string
	ctx              context.Context
	cancelFunc       context.CancelFunc
	httpClient       *http.Client
}

func MustNewAgent(opts ...Option) *Agent {
	a := &Agent{
		log:        &logger.DefaultLogger,
		ctx:        config.DefaultCtx(),
		httpClient: &http.Client{Timeout: callbackTimeout},
	}

	for _, o := range opts {
		o(a)
	}

	if a.callbackEndpoint == "" || a.watchDir == "" {
		a.log.Fatal().Msg("callbackEnpoint and watchDir is required for agent")
	}

	if a.cancelFunc == nil {
		ctx, cancel := config.DefaultCtxRootSysNotify(a.ctx)
		a.ctx = ctx
		a.cancelFunc = cancel
	}

	return a
}

func WithLogger(logger *zerolog.Logger) Option {
	return func(a *Agent) {
		a.log = logger
	}
}

func WithCallbackEndpoint(endpoint string) Option {
	return func(a *Agent) {
		a.callbackEndpoint = endpoint
	}
}

func WithWatchDir(dir string) Option {
	return func(a *Agent) {
		a.watchDir = dir
	}
}

func WithContext(ctx context.Context) Option {
	return func(a *Agent) {
		a.ctx = ctx
	}
}

func WithCancelFunc(cancel context.CancelFunc) Option {
	return func(a *Agent) {
		a.cancelFunc = cancel
	}
}

// RunAsync запускает цикл наблюдения в отдельной горутине.
// Возвращаемый канал закрывается после полного завершения цикла —
// по нему main ожидает очистку ресурсов при graceful shutdown.
func (a *Agent) RunAsync() <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		if err := a.run(); err != nil && a.ctx.Err() == nil {
			a.log.Error().Err(err).Msg("agent stopped with error")
			// Разблокируем main: без этого процесс живёт бесконечно уже без
			// наблюдения, а деплой по pgrep считает его работающим и не
			// переустанавливает.
			a.cancelFunc()
		}
	}()

	return done
}

func (a *Agent) run() error {
	fd, err := unix.FanotifyInit(unix.FAN_CLASS_NOTIF|unix.FAN_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("fanotify init (проверить capabilities контейнера, нужна CAP_SYS_ADMIN): %w", err)
	}
	defer unix.Close(fd)

	// Каталог наблюдения может ещё не существовать: агент стартует на свежем
	// хосте раньше загрузки первого скрипта, а fanotify mark по
	// несуществующему пути возвращает ENOENT.
	if err := os.MkdirAll(a.watchDir, 0o755); err != nil {
		return fmt.Errorf("create watch dir %s: %w", a.watchDir, err)
	}

	// Аргументы fanotify_mark(fd, flags, mask, dirfd, path) строго
	// позиционные: flags — тип операции над mark (FAN_MARK_ADD), mask —
	// сами события. Смешивать их нельзя: события в flags дают EINVAL.
	// Mark на каталог + FAN_EVENT_ON_CHILD даёт события по файлам,
	// лежащим в нём непосредственно, из какого бы cwd к ним ни обращались.
	if err := unix.FanotifyMark(fd,
		unix.FAN_MARK_ADD,
		unix.FAN_OPEN|unix.FAN_OPEN_EXEC|unix.FAN_EVENT_ON_CHILD,
		unix.AT_FDCWD, a.watchDir); err != nil {
		return fmt.Errorf("fanotify mark %s: %w", a.watchDir, err)
	}

	// Чтение событий — блокирующий unix.Read, поэтому отдельная горутина.
	// eventsFan закрытие fd при отмене контекста разблокирует Read ошибкой EBADF.
	eventsFan := make(chan fanEvent, 64)
	go a.readLoop(fd, eventsFan)

	a.log.Info().Str("watch_dir", a.watchDir).Str("callback", a.callbackEndpoint).Msg("agent started")

	for {
		select {
		case <-a.ctx.Done():
			// Закрываем fd — readLoop выйдет из блокированного Read и завершится.
			unix.Close(fd)
			a.log.Info().Msg("agent stopped")
			return nil
		case ev, ok := <-eventsFan:
			if !ok {
				return nil
			}
			a.handleEvent(ev)
		}
	}
}

func (a *Agent) readLoop(fd int, out chan<- fanEvent) {
	defer close(out)

	buf := make([]byte, 8192)
	for {
		n, err := unix.Read(fd, buf)
		if err != nil {
			// EBADF — fd закрыт при отмене контекста, это штатное завершение.
			if a.ctx.Err() != nil {
				return
			}
			a.log.Error().Err(err).Msg("read fanotify events failed")
			return
		}

		// Один Read возвращает пачку событий, слепленных подряд;
		// шаг между ними — event_len из самих метаданных.
		for off := 0; off+fanotifyMetaSize <= n; {
			eventLen := int(binary.NativeEndian.Uint32(buf[off:]))
			// Повреждённая длина означала бы выход за границы буфера — такой
			// пачке событий доверять нельзя, отбрасываем остаток.
			if eventLen < fanotifyMetaSize {
				break
			}

			mask := binary.NativeEndian.Uint64(buf[off+8:])
			fileFD := int(int32(binary.NativeEndian.Uint32(buf[off+16:])))
			pid := int(int32(binary.NativeEndian.Uint32(buf[off+20:])))

			out <- newFanEvent(mask, fileFD, pid)
			// Каждый event несёт собственный fd открытого файла — закрываем после разбора.
			unix.Close(fileFD)

			off += eventLen
		}
	}
}

func (a *Agent) handleEvent(ev fanEvent) {
	// Фильтр: наблюдаем только скрипты, размещённые сервисом (префикс общий с deployer).
	if !strings.HasPrefix(ev.scriptName(), entity.ScriptPrefix) {
		return
	}

	payload := entity.Event{
		User:       ev.user,
		ScriptPath: ev.path,
		Action:     ev.action,
		Time:       time.Now().UTC(),
	}

	if err := a.sendCallback(payload); err != nil {
		a.log.Error().Err(err).Str("script", payload.ScriptPath).Msg("send callback failed")
	}
}

func (a *Agent) sendCallback(p entity.Event) error {
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal callback payload: %w", err)
	}

	// callback_enpoint в конфиге — полный URL эндпоинта callback,
	// включая префикс роутера сервиса (service/version).
	resp, err := a.httpClient.Post(a.callbackEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("post callback: %w", err)
	}
	defer resp.Body.Close()

	// Любой 2xx считается успехом: успешный приём не обязан быть строго 200.
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("callback endpoint returned status %d", resp.StatusCode)
	}
	return nil
}

// processUser определяет имя пользователя по pid: эффективный uid берётся
// из /proc/<pid>/status, uid переводится в имя через системную базу (/etc/passwd).
func processUser(pid int) string {
	raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return "unknown"
	}

	// SplitSeq вместо Split: итератор не аллоцирует слайс строк целиком,
	// нужная строка Uid: обычно в начале файла.
	for line := range strings.SplitSeq(string(raw), "\n") {
		if !strings.HasPrefix(line, "Uid:") {
			continue
		}
		// Формат: "Uid:\t<real>\t<effective>\t<saved>\t<fs>"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return "unknown"
		}
		if u, err := user.LookupId(fields[1]); err == nil {
			return u.Username
		}
		return fields[1] // uid без записи в passwd — лучше число, чем "unknown"
	}
	return "unknown"
}
