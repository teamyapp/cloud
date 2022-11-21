package backoff

import (
	"time"
)

type BackOff interface {
	OnSuccess()
	OnFailure()
	Delay() time.Duration
}
