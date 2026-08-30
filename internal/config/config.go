package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env           string `yaml:"env" env-default:"prod"`
	HTTPServer    `yaml:"http_server"`
	DBUser        string `env:"DB_USER"`
	DBPassword    string `env:"DB_PASSWORD"`
	DBHost        string `env:"DB_HOST"`
	DBPort        int    `env:"DB_PORT"`
	DBName        string `env:"DB_NAME"`
	ParseTime     bool   `yaml:"parse_time"`
	Charset       string `yaml:"charset"`
	MigrationPath string `yaml:"migration_path" env:"MIGRATION_PATH"`

	AdminLogin string `yaml:"admin_login" env:"ADMIN_LOGIN"`
	AdminPass  string `yaml:"admin_pass" env:"ADMIN_PASS"`

	LDAPConfig `yaml:"ldap"`
	JWTConfig  `yaml:"jwt"`

	AuthConfig `yaml:"auth"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:4001"`
	Timeout     time.Duration `yaml:"timeout"  env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout"  env-default:"60s"`
}

type LDAPConfig struct {
	URL             string `yaml:"url"`
	AdminDN         string `yaml:"admin_dn"`
	AdminPassword   string `yaml:"admin_password"`
	BaseDN          string `yaml:"base_dn"`
	UserSearchBase  string `yaml:"user_search_base"`
	GroupSearchBase string `yaml:"group_search_base"`
	UserFilter      string `yaml:"user_filter"`
	GroupFilter     string `yaml:"group_filter"`
}

type JWTConfig struct {
	Secret          string `yaml:"secret" env:"JWT_SECRET"`
	ExpirationHours int    `yaml:"expiration_hours"`
}

type AuthConfig struct {
	Provider        string   `yaml:"provider"`
	AdminPermission []string `yaml:"admin_permission"`
}

func MustConfig() *Config {

	var cfg Config

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/local.yaml"
	}

	if err := godotenv.Load(); err != nil {
		log.Println("godotenv:", err)
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read env: %s", err)
	}

	log.Printf("config loaded: env=%s db_host=%s db_name=%s http=%s", cfg.Env, cfg.DBHost, cfg.DBName, cfg.HTTPServer.Address)

	return &cfg
}
