package num

import (
	"time"
)

func Min[Num int | uint64 | float32 | float64 | time.Duration](num1 Num, num2 Num) Num {
	if num1 < num2 {
		return num1
	} else {
		return num2
	}
}

func Max[Num int | uint64 | float32 | float64 | time.Duration](num1 Num, num2 Num) Num {
	if num1 > num2 {
		return num1
	} else {
		return num2
	}
}
