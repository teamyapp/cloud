package config

import (
	"log"
	"time"

	"github.com/teamyapp/one/config"
)

type Config struct {
	AccessTokenTTL     time.Duration `envconfig:"ACCESS_TOKEN_TTL" default:""`
	SignInTimeOut      time.Duration `envconfig:"SIGN_IN_TIMEOUT" default:""`
	JWTSigningKey      string        `envconfig:"JWT_SIGNING_KEY" default:""`
	GoogleClientID     string        `envconfig:"GOOGLE_CLIENT_ID" default:""`
	GoogleClientSecret string        `envconfig:"GOOGLE_CLIENT_SECRET" default:""`
	IDRangeLength      int           `envconfig:"ID_RANGE_LENGTH" default:"100"`
	WebAPIBaseURL      string        `envconfig:"WEB_API_BASE_URL" default:""`
	WebAPIPort         int           `envconfig:"WEB_API_PORT" default:"9500"`
	GRPCAPIPort        int           `envconfig:"GRPC_API_PORT" default:"9600"`
}

func FromEnv() (Config, error) {
	cfg := Config{}
	err := config.FromEnv(&cfg)
	if err != nil {
		log.Println(err)
		return Config{}, err
	}
	return cfg, nil
}
