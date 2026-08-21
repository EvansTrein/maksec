//go:build linux

package app

import (
	"os"
	"strconv"
	"strings"

	"maksec/internal/entity"

	"golang.org/x/sys/unix"
)

// fanEvent — одно событие fanotify после разбора метаданных ядра.
type fanEvent struct {
	path   string        // полный путь файла, получен через /proc/self/fd/N
	action entity.Action // entity.ActionOpen или entity.ActionExecute
	user   string        // пользователь процесса — разрешается сразу, см. newFanEvent
	pid    int           // pid процесса, совершившего действие
}

// newFanEvent собирает событие из полей метаданных fanotify, распарсенных
// из буфера чтения (см. readLoop в app.go).
// Путь и пользователь разрешаются здесь же, синхронно с разбором пачки:
// отправка callback — блокирующая, и пока уходит первое событие пачки,
// короткоживущий процесс (скрипт из echo/date) успевает завершиться,
// после чего /proc/<pid> исчезает и пользователь становится "unknown".
func newFanEvent(mask uint64, fd int, pid int) fanEvent {
	ev := fanEvent{pid: pid, user: processUser(pid)}

	switch {
	case mask&unix.FAN_OPEN_EXEC != 0:
		ev.action = entity.ActionExecute
	case mask&unix.FAN_OPEN != 0:
		ev.action = entity.ActionOpen
	}

	// Ядро отдаёт не путь, а уже открытый fd файла; реальный путь — симлинк в /proc.
	if link, err := os.Readlink("/proc/self/fd/" + strconv.Itoa(fd)); err == nil {
		ev.path = link
	}
	return ev
}

func (e fanEvent) scriptName() string {
	if i := strings.LastIndexByte(e.path, '/'); i >= 0 {
		return e.path[i+1:]
	}
	return e.path
}
