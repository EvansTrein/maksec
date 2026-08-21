//go:build linux

package main

import (
	_ "embed"
	"flag"
	"os"

	"maksec/cmd/agent/app"
	"maksec/internal/config"
	"maksec/pkg/logger"
)

//go:embed config.yaml
var defaultConfig []byte

func main() {
	cfgFile := flag.String("c", "config.yaml", "config file path")
	callbackURL := flag.String("callback", "", "callback endpoint URL (overrides config)")
	watchDir := flag.String("watch-dir", "", "scripts watch directory (overrides config)")
	flag.Parse()

	cfg, err := loadConfig(*cfgFile)
	if err != nil {
		logger.DefaultLogger.Fatal().Err(err).Msg("failed to load config file")
	}

	// Сервис при деплое передаёт актуальный адрес callback флагами —
	// бинарник агента не зависит от сетевой топологии места запуска.
	if *callbackURL != "" {
		cfg.Agent.CallbackEnpoint = *callbackURL
	}
	if *watchDir != "" {
		cfg.Agent.WatchDir = *watchDir
	}

	log := logger.NewLogger("agent", cfg.Logger.IsDebug, os.Stdout)

	rootCtx, rootCancelFunc := config.DefaultCtxRootSysNotify(config.DefaultCtx())
	defer rootCancelFunc()

	agent := app.MustNewAgent(
		app.WithLogger(log),
		app.WithCallbackEndpoint(cfg.Agent.CallbackEnpoint),
		app.WithWatchDir(cfg.Agent.WatchDir),
		app.WithContext(rootCtx),
		app.WithCancelFunc(rootCancelFunc),
	)

	// Канал done закрывается после полного завершения цикла наблюдения —
	// ждём его, чтобы не оборвать отправку callback при получении сигнала.
	done := agent.RunAsync()

	<-rootCtx.Done()
	log.Info().Msg("stopping agent")
	<-done
	log.Info().Msg("agent exited")
}

func loadConfig(file string) (*config.Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		data = defaultConfig
	}
	return config.LoadFromBytes(data)
}
