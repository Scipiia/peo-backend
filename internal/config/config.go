package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"prod"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
	DBUser      string `yaml:"db_user" env-required:"true"`
	DBPassword  string `yaml:"db_password" env-required:"false"`
	DBHost      string `yaml:"db_host" env-default:"localhost"`
	DBPort      int    `yaml:"db_port" env-default:"3306"`
	DBName      string `yaml:"db_name" env-required:"true"`
	ParseTime   bool   `yaml:"parse_time" env-required:"true"`
	Charset     string `yaml:"charset"`

	AdminLogin string `yaml:"admin_login"`
	AdminPass  string `yaml:"admin_pass"`

	LDAPConfig `yaml:"ldap"`
	JWTConfig  `yaml:"jwt"`

	AuthConfig `yaml:"auth"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:4001"`
	Timeout     time.Duration `yaml:"timeout"  env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout"  env-default:"60s"`
	//User        string        `yaml:"user" env-required:"true"`
	//Password    string        `yaml:"password" env-required:"true"`
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
	Secret          string `yaml:"secret"`
	ExpirationHours int    `yaml:"expiration_hours"`
}

type AuthConfig struct {
	Provider        string   `yaml:"provider"`
	AdminPermission []string `yaml:"admin_permission"`
}

func MustConfig() *Config {
	//configPath := os.Getenv("CONFIG_PATH")
	//if configPath == "" {
	//	log.Fatal("CONFIG_PATH is not set")
	//}
	////log.Println(configPath)
	//
	//if _, err := os.Stat(configPath); os.IsNotExist(err) {
	//	log.Fatalf("config file does not exist: %s", configPath)
	//}

	var cfg Config
	a := "./config/local.yaml"

	if err := cleanenv.ReadConfig(a, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
