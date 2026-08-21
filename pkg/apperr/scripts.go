package apperr

import "errors"

var (
	ErrTemplateNotFound = errors.New("scripts.template.notFound")
	ErrSSHAuthFailed    = errors.New("scripts.ssh.authFailed")
)
