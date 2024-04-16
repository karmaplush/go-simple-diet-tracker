package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string        `yaml:"env"          env-default:"local"`
	StoragePath string        `yaml:"storage_path"                     env-required:"true"`
	HttpServer  HttpServer    `yaml:"http_server"`
	Clients     ClientsConfig `yaml:"clients"`
	AppSecret   string        `yaml:"app_secret"                       env-required:"true" env:"APP_SECRET"`
	AppId       int32         `yaml:"app_id"                           env-required:"true" env:"APP_ID"`
}

type HttpServer struct {
	Address     string        `yaml:"address"      env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout"      env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Client struct {
	Address      string        `yaml:"address"`
	Timeout      time.Duration `yaml:"timeout"`
	RetriesCount int           `yaml:"retries_count"`
}

type ClientsConfig struct {
	GRPCAuth Client `yaml:"grpc_auth"`
}

func MustLoad() *Config {

	path := mustFetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file is not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg

}

func mustFetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	if res == "" {
		panic("--config flag not set and CONFIG_PATH env not found")
	}

	return res
}
