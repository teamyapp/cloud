package runner

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type Service interface {
	Start(runner *ServiceRunner) *errs.Error
}
