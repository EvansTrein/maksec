package main

import (
	"flag"
	"os"

	"maksec/cmd/scripts/api"
	"maksec/internal/adapter/httpserver"
	"maksec/internal/assets"
	"maksec/internal/config"
	"maksec/internal/deploy"
	"maksec/internal/repo"
	"maksec/internal/usecase"
	"maksec/pkg/logger"
	"maksec/pkg/postgres"
	"maksec/pkg/rest"
)

func main() {
	var (
		err error

		db *postgres.Database

		deployer *deploy.Deployer

		scriptsRepo *repo.Scripts
		eventsRepo  *repo.Events

		scriptsUC        *usecase.Scripts
		callbackUC       *usecase.Callback
		base             *httpserver.HandlerBase
		handlersScripts  *httpserver.HandlerScripts
		handlersCallback *httpserver.HandlerCallback

		apiScripts  *api.ScriptsActiveHandlers
		apiCallback *api.CallbackActiveHandlers

		scriptsRouter  *rest.Router
		callbackRouter *rest.Router

		scriptsServer  *rest.Server
		callbackServer *rest.Server
	)

	cfgFile := flag.String("c", "config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		logger.DefaultLogger.Fatal().Err(err).Msg("failed to load config file")
	}

	log := logger.NewLogger(cfg.Scripts.Name, cfg.Logger.IsDebug, os.Stdout)

	templates := assets.Templates()
	agentBin, err := assets.AgentBinary()
	if err != nil {
		log.Fatal().Err(err).Msg("agent binary not found in embed (run make build-agent and rebuild)")
	}

	rootCtx, rootCancel := config.DefaultCtxRootSysNotify(config.DefaultCtx())
	defer rootCancel()

	// Init infrastructure
	{
		// Init Postgres
		db, err = postgres.New(cfg.DsnPostgres(), postgres.WithContext(rootCtx))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to database")
		}
	}

	{ // Init helpers
		deployer = deploy.NewDeployer(log, cfg.Agent.CallbackEnpoint, cfg.Agent.WatchDir)
	}

	{ // Init repo
		scriptsRepo = repo.MustNewScripts(log, db)
		eventsRepo = repo.MustNewEvents(log, db)
	}

	{ // Init usecase
		scriptsUC = usecase.NewScripts(
			log,
			scriptsRepo,
			deployer,
			templates,
			agentBin,
		)
		callbackUC = usecase.NewCallback(log, eventsRepo)
	}

	{ // Init handler
		base = httpserver.NewHandlerBase(log)
		handlersScripts = httpserver.NewHandlerScripts(base, scriptsUC)
		handlersCallback = httpserver.NewHandlerCallback(base, callbackUC)
	}

	{ // Init api
		apiScripts = api.NewScriptsActiveHandlers(handlersScripts)
		apiCallback = api.NewCallbackActiveHandlers(handlersCallback)
	}

	// Init servers
	{
		// Scripts
		scriptsRouter = api.InitRouterScripts(cfg.Scripts.Version, cfg.Scripts.Name, apiScripts)
		scriptsServer = rest.MustNewServer(
			rest.WithHost(cfg.Scripts.Host),
			rest.WithPort(cfg.Scripts.Port),
			rest.WithRouter(scriptsRouter),
			rest.WithChecker(db.Ping),
			rest.WithLogger(log),
			rest.WithContext(rootCtx),
		)
		go scriptsServer.MustStart()

		// Callback
		callbackRouter = api.InitRouterCallback(cfg.Scripts.Version, cfg.Scripts.Name, apiCallback)
		callbackServer = rest.MustNewServer(
			rest.WithHost(cfg.Scripts.Host),
			rest.WithPort(cfg.Scripts.AgentCallbackPort),
			rest.WithRouter(callbackRouter),
			rest.WithChecker(db.Ping),
			rest.WithLogger(log),
			rest.WithContext(rootCtx),
		)
		go callbackServer.MustStart()
	}

	log.Info().Str("callback", cfg.Agent.CallbackEnpoint).Str("watch_dir", cfg.Agent.WatchDir).Msg("service started")

	<-rootCtx.Done()
	log.Info().Msg("shutdown signal received")

	if err := scriptsServer.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to stop primary server")
	}

	if err := callbackServer.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to stop callback server")
	}

	if err := db.Close(); err != nil {
		log.Error().Err(err).Msg("failed to stop postgres")
	}

	log.Info().Msg("service stopped")
}
