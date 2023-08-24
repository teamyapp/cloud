package randgentest

type StubRanGen struct {
	randomInts         []int
	nextRandomIntIndex int
}

func (s *StubRanGen) RandomInt(i int) int {
	if s.nextRandomIntIndex == len(s.randomInts) {
		s.nextRandomIntIndex = 0
		return 0
	}

	value := s.randomInts[s.nextRandomIntIndex]
	s.nextRandomIntIndex++
	return value
}

func NewStubRanGen(randomInts []int) *StubRanGen {
	return &StubRanGen{
		randomInts:         randomInts,
		nextRandomIntIndex: 0,
	}
}
