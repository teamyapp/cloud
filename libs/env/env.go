package env

type Environment string

const (
	ProductionEnv  Environment = "production"
	StagingEnv     Environment = "staging"
	TestingEnv     Environment = "testing"
	DevelopmentEnv Environment = "development"
)
