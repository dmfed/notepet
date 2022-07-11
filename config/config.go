package config

const (
	AuthNone  AuthType = "none"
	AuthToken AuthType = "token"
	AuthBasic AuthType = "basic"
)

type AuthType string

type ClientConfig struct {
	CLI  CliConfig `yaml:"cli"`
	Conn Conn      `yaml:"conn"`
	Auth Auth      `yaml:"auth"`
}

type ServerConfig struct {
	Conn Conn `yaml:"conn"`
}

type CliConfig struct {
	Editor  string `yaml:"editor"`
	Verbose bool   `yaml:"verbose"`
	Color   Color  `yaml:"color"`
}

type Color struct {
	Header string `yaml:"header"`
	Sticky string `yaml:"sticky"`
	Tags   string `yaml:"tags"`
	Body   string `yaml:"body"`
	Error  string `yaml:"error"`
}

type Conn struct {
	IP   string `yaml:"ip"`
	Port string `yaml:"port"`
	Path string `yaml:"path"`
	TLS  bool   `yaml:"tls"`
}

type Auth struct {
	Type     AuthType `yaml:"type"`
	User     string   `yaml:"user"`
	Password string   `yaml:"pass"`
}
