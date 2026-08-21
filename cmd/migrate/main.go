package main

import (
	"flag"
	"maksec/internal/config"
	"maksec/pkg/logger"
	db "maksec/pkg/postgres"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	c := flag.String("c", "config.yaml", "path to config file")
	up := flag.Bool("u", false, "Migrate Up")
	down := flag.Bool("d", false, "Migrate Down")
	step := flag.Int("s", 0, "Migrate step")
	force := flag.Int("f", 0, "Force version")
	flag.Parse()

	log := logger.NewLogger("migrate", false, os.Stdout)

	cfg, err := config.Load(*c)
	if err != nil {
		log.Fatal().Err(err).Msg("failedfailed to load config")
	}

	db, err := db.New(cfg.DsnPostgres())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	driver, err := postgres.WithInstance(db.SqlConnection(), &postgres.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init database driver")
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create database instance")
	}

	if force != nil && *force > 0 {
		if err = m.Force(*force); err != nil {
			log.Error().Err(err).Msg("error force version")
		}
	}
	if up != nil && *up {
		log.Info().Msg("start migrate up")
		if err = m.Up(); err != nil {
			log.Error().Err(err).Msg("error migrate up")
		}
	} else if down != nil && *down {
		log.Info().Msg("start migrate down")
		if err = m.Down(); err != nil {
			log.Error().Err(err).Msg("error migrate down")
		}
	} else if step != nil && *step != 0 {
		log.Info().Msg("start migrate step")
		if err = m.Steps(*step); err != nil {
			log.Error().Err(err).Msg("error migrate step")
		}
	}

	v, dirty, err := m.Version()
	if err != nil {
		log.Error().Err(err).Msg("error get migration version")
	} else {
		log.Info().Uint("Version", v).Bool("Dirty", dirty).Msg("migration current version")
	}
}
