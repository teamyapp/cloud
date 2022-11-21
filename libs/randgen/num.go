package randgen

import "math/rand"

type BuiltinRanGen struct {
}

func (b BuiltinRanGen) RandomInt(i int) int {
	return rand.Intn(i)
}

func NewBuiltinRanGen() BuiltinRanGen {
	return BuiltinRanGen{}
}
