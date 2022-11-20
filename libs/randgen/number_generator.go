package randgen

type RandomNumberGenerator interface {
	RandomInt(max int) int
}
