package app

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Log struct {
		Level string `env-required:"true" env:"LOG_LEVEL"`
	}

	TG struct {
		Token string `env-required:"true" env:"TG_TOKEN"`
	}
}

func NewConfig() (*Config, error) {
	cfg := new(Config)

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
