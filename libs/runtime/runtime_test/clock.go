package runtime_test

import "time"

type TestClock struct {
	time time.Time
}

func (t *TestClock) Now() time.Time {
	return t.time
}

func (t *TestClock) SetTime(time time.Time) {
	t.time = time
}

func NewTestClock(time time.Time) TestClock {
	return TestClock{
		time: time,
	}
}
