package entity

import (
	"time"
)

type Action string

const (
	ActionOpen    Action = "open"
	ActionExecute Action = "execute"
)

type Event struct {
	User       string    `json:"user"`
	ScriptPath string    `json:"script"`
	Action     Action    `json:"action"`
	Time       time.Time `json:"time"`
}

type EventRow struct {
	ID         int64     `db:"id"`
	ScriptPath string    `db:"script_path"`
	AgentUser  string    `db:"agent_user"`
	Action     Action    `db:"action"`
	EventTime  time.Time `db:"event_time"`
	CreatedAt  time.Time `db:"created_at"`
}
