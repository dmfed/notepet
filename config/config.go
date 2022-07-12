package config

const (
	StorageSQLite   StorageType = "sqlite"
	StoragePostgres StorageType = "postgres"
	StorageNetwork  StorageType = "network"
)

type AuthType string

type StorageType string

type ClientConfig struct {
	Storage StorageConfig `yaml:"storage"`
	Auth    Auth          `yaml:"auth"`
	CLI     CLIConfig     `yaml:"cli"`
	Color   ColorConfig   `yaml:"color"`
}

type ServerConfig struct {
	Storage StorageConfig `yaml:"storage"`
	URL     string        `yaml:"url"`
}

type StorageConfig struct {
	Type   StorageType `yaml:"type"`   // type of storage to use: json, sqlite, postgresql or network
	Source string      `yaml:"source"` // filename for JSON and SQLite, dsn for Postgres or URL for network
}

type CLIConfig struct {
	Editor  string `yaml:"editor"`
	Verbose bool   `yaml:"verbose"`
}

type ColorConfig struct {
	Enabled bool   `yaml:"enabled"`
	Header  string `yaml:"header"`
	Sticky  string `yaml:"sticky"`
	Tags    string `yaml:"tags"`
	Body    string `yaml:"body"`
	Error   string `yaml:"error"`
}

type Auth struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}
