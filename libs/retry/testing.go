package retry

import "time"

const shortDelay = 200 * time.Millisecond
const longDelay = 250 * time.Millisecond
const randomOffset = 1 * time.Millisecond
