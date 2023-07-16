package runtime

import "time"

type Clock interface {
	Now() time.Time
	After(d time.Duration, id uint64) <-chan time.Time
}
