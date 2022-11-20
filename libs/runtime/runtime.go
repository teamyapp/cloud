package runtime

import "time"

type Runtime interface {
	Sleep(duration time.Duration)
}
