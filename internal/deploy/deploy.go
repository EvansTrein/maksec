package deploy

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"maksec/internal/entity"
	"maksec/pkg/apperr"

	"github.com/pkg/sftp"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/ssh"
)

const (
	agentDir   = "/opt/maksec"
	scriptsDir = "/var/lib/maksec/scripts"

	sshConnectTimeout = 10 * time.Second

	sftpCreateOrTrunc = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
)

type Deployer struct {
	log         *zerolog.Logger
	callbackURL string
	watchDir    string
}

// NewDeployer принимает адреса из конфига сервиса: они передаются агенту
// флагами при запуске, чтобы бинарник агента не зависел от сетевой топологии.
func NewDeployer(log *zerolog.Logger, callbackURL, watchDir string) *Deployer {
	return &Deployer{
		log:         log,
		callbackURL: callbackURL,
		watchDir:    watchDir,
	}
}

// Connect устанавливает SSH-сессию (пароль или ключ — что заполнено в script)
// и возвращает клиента. Ответственность за Close — на вызывающем.
// ssh.Dial не принимает ctx, поэтому TCP-соединение строится через
// net.Dialer с контекстом, а SSH-хендшейк — поверх готового conn.
func (d *Deployer) Connect(ctx context.Context, script *entity.Script) (*ssh.Client, error) {
	log := d.log.With().Str("operation", "ssh-connect").Str("host", script.Host).Logger()

	cfg := &ssh.ClientConfig{
		User:            script.SSHUser,
		Auth:            authMethods(script.Password),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // тестовый стенд; для продакшена — known_hosts
		Timeout:         sshConnectTimeout,
	}

	addr := net.JoinHostPort(script.Host, "22")

	netDialer := &net.Dialer{Timeout: sshConnectTimeout}
	conn, err := netDialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("tcp dial %s: %w", addr, err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, cfg)
	if err != nil {
		conn.Close() // nolint: errcheck
		if strings.Contains(err.Error(), "unable to authenticate") {
			log.Warn().Err(err).Msg("auth ssh failed")
			return nil, apperr.ErrSSHAuthFailed
		}
		return nil, fmt.Errorf("ssh handshake %s: %w", addr, err)
	}

	client := ssh.NewClient(sshConn, chans, reqs)
	log.Debug().Msg("connected")
	return client, nil
}

// authMethods строит список методов аутентификации: пароль, если задан.
// Слайс позволяет позже добавить publickey без изменения вызывающего кода.
func authMethods(password string) []ssh.AuthMethod {
	methods := make([]ssh.AuthMethod, 0, 1)
	if password != "" {
		methods = append(methods, ssh.Password(password))
	}
	return methods
}

// UploadScript кладёт скрипт на хост и возвращает полный путь к нему.
func (d *Deployer) UploadScript(ctx context.Context, client *ssh.Client, template, content string) (string, error) {
	log := d.log.With().Str("operation", "upload-script").Logger()
	_ = ctx

	path, err := d.upload(client, scriptsDir, entity.ScriptPrefix+template+".sh", content)
	if err != nil {
		return "", err
	}
	log.Debug().Str("path", path).Msg("script uploaded")
	return path, nil
}

// UploadAgent кладёт бинарник агента на хост и запускает его, если он ещё
// не работает. Повторный вызов на том же хосте не перезапускает агента —
// наблюдение продолжается. Адрес callback и каталог наблюдения передаются
// агенту флагами при запуске.
func (d *Deployer) UploadAgent(ctx context.Context, client *ssh.Client, agentBin []byte) error {
	log := d.log.With().Str("operation", "upload-agent").Logger()

	// Если агент уже работает — не трогаем ни процесс, ни бинарник:
	// запись в исполняемый файл работающего процесса заканчивается ETXTBSY
	// (sftp отдаёт её как безликий SSH_FX_FAILURE).
	// Паттерн "^...( |$)" матчит cmdline запуска агента с любыми аргументами;
	// якорь ^ исключает самоматчинг sh -c, в командной строке которого
	// есть этот же путь.
	if err := d.run(ctx, client, fmt.Sprintf("pgrep -f '^%s( |$)' >/dev/null", agentBinPath())); err == nil {
		log.Debug().Msg("agent already running, skip reinstall")
		return nil
	}

	if _, err := d.upload(client, agentDir, "agent", string(agentBin)); err != nil {
		return err
	}

	// Адрес callback и каталог наблюдения передаются флагами: бинарник
	// агента не зависит от топологии сети, в которой запущен сервис.
	// Вывод — в stdout контейнера (/proc/1/fd/1 — stdout PID 1), логи видны
	// в docker logs; запись в чужой fd требует root — деплой идёт под ним.
	run := fmt.Sprintf("nohup %s -callback '%s' -watch-dir '%s' >/proc/1/fd/1 2>&1 &",
		agentBinPath(), d.callbackURL, d.watchDir)
	if err := d.run(ctx, client, run); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}
	log.Debug().Str("path", agentBinPath()).Msg("agent uploaded and running")
	return nil
}

func agentBinPath() string { return agentDir + "/agent" }

// upload пишет файл в dir с правом исполнения и возвращает его путь.
func (d *Deployer) upload(client *ssh.Client, dir, name, content string) (string, error) {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return "", fmt.Errorf("sftp client: %w", err)
	}
	defer sftpClient.Close() // nolint: errcheck

	if err := sftpClient.MkdirAll(dir); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", dir, err)
	}

	path := strings.TrimSuffix(dir, "/") + "/" + name

	file, err := sftpClient.OpenFile(path, sftpCreateOrTrunc)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close() // nolint: errcheck

	if _, err := io.WriteString(file, content); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}

	if err := sftpClient.Chmod(path, 0o755); err != nil {
		return "", fmt.Errorf("chmod %s: %w", path, err)
	}
	return path, nil
}

// run выполняет одну shell-команду на хосте, respecting ctx.
func (d *Deployer) run(ctx context.Context, client *ssh.Client, command string) error {
	sess, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer sess.Close() // nolint: errcheck

	done := make(chan error, 1)
	go func() { done <- sess.Run(command) }()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
