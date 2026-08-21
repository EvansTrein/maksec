package usecase

import (
	"context"
	"fmt"
	"io/fs"

	"maksec/internal/dto"
	"maksec/internal/entity"
	"maksec/pkg/apperr"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/ssh"
)

type IRepoScripts interface {
	Create(ctx context.Context, script *entity.Script) (*entity.Script, error)
}

type IDepoyer interface {
	Connect(ctx context.Context, script *entity.Script) (*ssh.Client, error)
	UploadAgent(ctx context.Context, client *ssh.Client, agentBin []byte) error
	UploadScript(ctx context.Context, client *ssh.Client, template string, content string) (string, error)
}

type Scripts struct {
	log       *zerolog.Logger
	repo      IRepoScripts
	deployer  IDepoyer
	templates fs.FS
	agentBin  []byte
}

func NewScripts(
	log *zerolog.Logger,
	repo IRepoScripts,
	deployer IDepoyer,
	templates fs.FS,
	agentBin []byte,
) *Scripts {
	return &Scripts{
		log:       log,
		repo:      repo,
		deployer:  deployer,
		templates: templates,
		agentBin:  agentBin,
	}
}

func (us *Scripts) Create(ctx context.Context, req *dto.ScriptsCreateRequest) (*dto.ScriptsCreateResponse, error) {
	log := us.log.With().Str("operation", "usecase-create").Logger()

	content, err := us.templateContent(req.Template)
	if err != nil {
		log.Error().Err(err).Str("template", req.Template).Msg("failed to get template")
		return nil, err
	}

	script, err := us.deployScript(ctx, req, content)
	if err != nil {
		log.Error().Err(err).Str("host", req.Host).Msg("failed to deploy script")
		return nil, err
	}

	saved, err := us.repo.Create(ctx, script)
	if err != nil {
		log.Error().Err(err).Msg("failed to save script in db")
		return nil, err
	}

	resp := &dto.ScriptsCreateResponse{Script: *saved}

	log.Debug().Str("path", saved.Path).Msg("success")
	return resp, nil
}

// templateContent возвращает содержимое шаблона из embed по имени из запроса.
func (us *Scripts) templateContent(name string) (string, error) {
	path := name + ".sh"
	if _, err := fs.Stat(us.templates, path); err != nil {
		return "", apperr.ErrTemplateNotFound
	}
	raw, err := fs.ReadFile(us.templates, path)
	if err != nil {
		return "", fmt.Errorf("read template %q: %w", name, err)
	}
	return string(raw), nil
}

// deployScript — SSH-часть: подключение, установка агента, загрузка скрипта.
// Возвращает сущность с заполненным Path.
func (us *Scripts) deployScript(ctx context.Context, req *dto.ScriptsCreateRequest, content string) (*entity.Script, error) {
	log := us.log.With().Str("operation", "deploy-script").Str("host", req.Host).Logger()

	script := &entity.Script{
		Host:     req.Host,
		SSHUser:  req.User,
		Template: req.Template,
		Password: req.Password,
	}

	client, err := us.connect(ctx, script)
	if err != nil {
		return nil, err
	}
	defer client.Close() // nolint: errcheck

	if err := us.installAgent(ctx, client); err != nil {
		return nil, err
	}

	script.Path, err = us.uploadScript(ctx, client, req.Template, content)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("path", script.Path).Msg("deployed")
	return script, nil
}

// connect подключается к хосту из script.
func (us *Scripts) connect(ctx context.Context, script *entity.Script) (*ssh.Client, error) {
	log := us.log.With().Str("operation", "ssh-connect").Str("host", script.Host).Logger()

	client, err := us.deployer.Connect(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", script.Host, err)
	}
	log.Debug().Msg("connected")
	return client, nil
}

// installAgent заливает бинарник агента из embed и запускает его.
func (us *Scripts) installAgent(ctx context.Context, client *ssh.Client) error {
	log := us.log.With().Str("operation", "install-agent").Logger()

	if err := us.deployer.UploadAgent(ctx, client, us.agentBin); err != nil {
		return fmt.Errorf("upload agent: %w", err)
	}
	log.Debug().Msg("agent installed")
	return nil
}

// uploadScript кладёт скрипт по шаблону на хост, возвращает путь.
func (us *Scripts) uploadScript(ctx context.Context, client *ssh.Client, template, content string) (string, error) {
	log := us.log.With().Str("operation", "upload-script").Logger()

	path, err := us.deployer.UploadScript(ctx, client, template, content)
	if err != nil {
		return "", fmt.Errorf("upload script: %w", err)
	}
	log.Debug().Str("path", path).Msg("script uploaded")
	return path, nil
}
