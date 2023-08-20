package testkit

import (
	"time"
)

type Config struct {
	GenUniqueNumberRangeSize uint64
	JWTSigningKey            string
	AccessTokenTTL           time.Duration
	WebAPIBaseURL            string
	GithubClientID           string
	GithubClientSecret       string
	GoogleClientID           string
	GoogleClientSecret       string
	SlackClientID            string
	SlackClientSecret        string
	WebServerPort            int
	GRPCServerPort           int
}
