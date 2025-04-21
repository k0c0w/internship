package config

import (
	"errors"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

const (
	MustBeYmlFileError                 string = "config must be in yaml format"
	FailedToReadConfigPrefixError      string = "failed to read config"
	FailedToUnMarshalConfigPrefixError string = "failed to unmarshal config"
)

type (
	Config struct {
		PostgresConfig `mapstructure:"postgres"`
		HTTPConfig     `mapstructure:"http-profile"`
		GRPCConfig     `mapstructure:"grpc-profile"`
		AuthConfig     `mapstructure:"auth"`
	}

	PostgresConfig struct {
		Host            string        `mapstructure:"host"`
		Port            int           `mapstructure:"port"`
		User            string        `mapstructure:"user"`
		Password        string        `mapstructure:"password"`
		DB              string        `mapstructure:"db"`
		ConnectAttempts int           `mapstructure:"connect-attempts"`
		ConnectTimeout  time.Duration `mapstructure:"connect-timeout"`
	}

	HTTPConfig struct {
		Host           string `mapstructure:"host"`
		Port           int    `mapstructure:"port"`
		IncludeSwagger bool   `mapstructure:"include-swagger"`
	}

	AuthConfig struct {
		JWTConfig `mapstructure:"jwt"`
	}

	JWTConfig struct {
		TokenTTL time.Duration `mapstructure:"tokenTTL"`
		Sign     string        `mapstructure:"sign"`
		Issuer   string        `mapstructure:"issuer"`
	}

	GRPCConfig struct {
		Port int `mapstructure:"port"`
	}
)

func InitConfig(yamlConfigPath string) (Config, error) {
	v := viper.New()

	extension := filepath.Ext(yamlConfigPath)
	if extension != ".yaml" && extension != ".yml" {
		return Config{}, errors.New(MustBeYmlFileError)
	}

	v.SetConfigFile(yamlConfigPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return Config{}, errors.Join(errors.New(FailedToReadConfigPrefixError), err)
	}

	v.AutomaticEnv()
	v.BindEnv("postgres.user", "POSTGRES_USER")
	v.BindEnv("postgres.password", "POSTGRES_PASSWORD")
	v.BindEnv("postgres.db", "POSTGRES_DB")
	v.BindEnv("auth.jwt.sign", "JWT_SIGN")

	cfg := Config{}

	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, errors.Join(errors.New(FailedToUnMarshalConfigPrefixError), err)
	}

	return cfg, nil
}
