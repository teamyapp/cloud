package randgen_test

type BuiltinRanGen struct {
	randomInts         []int
	nextRandomIntIndex int
}

func (b BuiltinRanGen) RandomInt(i int) int {
	if b.nextRandomIntIndex == len(b.randomInts) {
		b.nextRandomIntIndex = 0
		return 0
	}

	value := b.randomInts[b.nextRandomIntIndex]
	b.nextRandomIntIndex++
	return value
}

func NewBuiltinRanGen(randomInts []int) BuiltinRanGen {
	return BuiltinRanGen{
		randomInts:         randomInts,
		nextRandomIntIndex: 0,
	}
}
