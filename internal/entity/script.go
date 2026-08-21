package entity

import "time"

// ScriptPrefix — префикс имён файлов, размещаемых сервисом на целевом хосте.
// Единый источник для deployer (именование файлов) и агента (фильтр событий),
// чтобы обе стороны всегда были согласованы.
const ScriptPrefix = "maksec_"

type Script struct {
	ID        int64     `json:"id" db:"id"`
	Host      string    `json:"host" db:"host"`
	SSHUser   string    `json:"ssh_user" db:"ssh_user"`
	Password  string    `json:"password" db:"password"`
	Template  string    `json:"template" db:"template"`
	Path      string    `json:"path" db:"path"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
