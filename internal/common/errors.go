package common

// arbitrary numbers that hopefully doesn't collide with any existing error codes
const (
	IOError int = -1923754812 - iota
	FormatError
	CodecError
)
