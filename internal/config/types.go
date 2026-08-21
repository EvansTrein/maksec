package config

type Logger struct {
	FilePath string `mapstructure:"file_path" default:"./logs"`
	IsDebug  bool   `mapstructure:"is_debug" default:"false"`
}

type Scripts struct {
	Name              string `mapstructure:"name" default:"scripts"`
	Version           string `mapstructure:"version" default:"v1"`
	Host              string `mapstructure:"host" default:"localhost"`
	Port              string `mapstructure:"port" default:"8080"`
	AgentCallbackPort string `mapstructure:"agent_callback_port" default:"8081"`
}

type Agent struct {
	Version         string `mapstructure:"version" default:"v1"`
	CallbackEnpoint string `mapstructure:"callback_enpoint" default:"http://localhost:8081/scripts/v1/callback"`
	WatchDir        string `mapstructure:"watch_dir" default:"./var/lib/maksec/scripts"`
}

type Postgres struct {
	Host     string `mapstructure:"host" default:"localhost"`
	Port     int    `mapstructure:"port" default:"5432"`
	Username string `mapstructure:"username" default:"admin"`
	Password string `mapstructure:"password" default:"123456"`
	Database string `mapstructure:"database" default:"maksec"`
	UseSsl   string `mapstructure:"use_ssl" default:"disable"`
}
