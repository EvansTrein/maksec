package usecase_test

import (
	"context"
	"testing"
	"time"

	"maksec/internal/entity"
	"maksec/internal/usecase"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRepoEvents — простой мок IRepoEvents: запоминает последнюю записанную
// строку и ошибку, которую нужно вернуть из Create.
type mockRepoEvents struct {
	saved *entity.EventRow
	err   error
}

func (m *mockRepoEvents) Create(_ context.Context, event *entity.EventRow) error {
	if m.err != nil {
		return m.err
	}
	m.saved = event
	return nil
}

func Test_CallbackExec(t *testing.T) {
	log := zerolog.Nop()

	eventTime := time.Date(2026, 8, 21, 12, 30, 15, 0, time.UTC)
	event := &entity.Event{
		User:       "root",
		ScriptPath: "/var/lib/maksec/scripts/maksec_template1.sh",
		Action:     entity.ActionExecute,
		Time:       eventTime,
	}

	mock := &mockRepoEvents{}
	uc := usecase.NewCallback(&log, mock)

	err := uc.Exec(context.Background(), event)
	require.NoError(t, err)

	require.NotNil(t, mock.saved, "event row must be passed to repo")
	assert.Equal(t, event.ScriptPath, mock.saved.ScriptPath)
	assert.Equal(t, event.User, mock.saved.AgentUser)
	assert.Equal(t, event.Action, mock.saved.Action)
	assert.Equal(t, event.Time, mock.saved.EventTime)
}
