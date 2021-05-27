package avformat

// #include <libavformat/avformat.h>
// #include <libavformat/avio.h>
import "C"

const (
	IOFlagWrite = C.AVIO_FLAG_WRITE
	IOFlagRead  = C.AVIO_FLAG_READ
)

const (
	NoFile = C.AVFMT_NOFILE
)
