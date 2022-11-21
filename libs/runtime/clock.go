package runtime

import "time"

type Clock interface {
	Now() time.Time
}
