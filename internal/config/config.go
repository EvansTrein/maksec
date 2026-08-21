package config

import (
	"fmt"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

type Config struct {
	Logger   Logger   `mapstructure:"logger" default:""`
	Scripts  Scripts  `mapstructure:"scripts" default:""`
	Agent    Agent    `mapstructure:"agent" default:""`
	Postgres Postgres `mapstructure:"postgres" default:""`
}

func Load(file string) (*Config, error) {
	c := newParser()
	if err := c.LoadFiles(file); err != nil {
		return nil, err
	}
	return decode(c)
}

// LoadFromBytes парсит конфиг из данных, а не с диска. Используется агентом,
// у которого default-конфиг зашит в бинарник через embed.
func LoadFromBytes(data []byte) (*Config, error) {
	c := newParser()
	if err := c.LoadSources("yaml", data); err != nil {
		return nil, err
	}
	return decode(c)
}

func newParser() *config.Config {
	return config.New("service", config.ParseEnv, config.ParseDefault, config.ParseTime).WithDriver(yaml.Driver)
}

func decode(c *config.Config) (*Config, error) {
	var cfg Config
	if err := c.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) DsnPostgres() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Postgres.Username,
		c.Postgres.Password,
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.Database,
		c.Postgres.UseSsl)
}
