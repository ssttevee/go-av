package av

// #include <libavformat/avio.h>
//
// extern int goavFileReadPacket(void *opaque, uint8_t *buf, int buf_size);
// extern int goavFileWritePacket(void *opaque, uint8_t *buf, int buf_size);
// extern int64_t goavFileSeek(void *opaque, int64_t offset, int whence);
import "C"
import (
	"io"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/ssttevee/go-av/avformat"
	"github.com/ssttevee/go-av/avutil"
)

type pinnedFile struct {
	f interface{}

	err error
}

func (f *pinnedFile) Read(p []byte) (n int, err error) {
	return f.f.(io.Reader).Read(p)
}

func (f *pinnedFile) Write(p []byte) (n int, err error) {
	return f.f.(io.Writer).Write(p)
}

func (f *pinnedFile) Seek(offset int64, whence int) (int64, error) {
	return f.f.(io.Seeker).Seek(offset, whence)
}

var pinnedFiles = map[pinType]*pinnedFile{}

func getPinnedFile(p unsafe.Pointer) io.ReadWriteSeeker {
	f, ok := pinnedFiles[pin(p)]
	if !ok {
		panic("pinned file not found")
	}

	return f
}

func returnPinnedFileError(p unsafe.Pointer, err error) C.int {
	pinnedFiles[pin(p)].err = err
	return C.int(errInternalIOError)
}

func wrapCBuf(buf *C.uint8_t, size C.int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(buf)),
		Len:  int(size),
		Cap:  int(size),
	}))
}

//export goavFileReadPacket
func goavFileReadPacket(p unsafe.Pointer, buf *C.uint8_t, bufSize C.int) C.int {
	n, err := getPinnedFile(p).Read(wrapCBuf(buf, bufSize))
	if err == io.EOF {
		return C.AVERROR_EOF
	} else if err != nil {
		return returnPinnedFileError(p, err)
	}

	return C.int(n)
}

//export goavFileWritePacket
func goavFileWritePacket(p unsafe.Pointer, buf *C.uint8_t, bufSize C.int) C.int {
	_, err := getPinnedFile(p).Write(wrapCBuf(buf, bufSize))
	if err != nil {
		return returnPinnedFileError(p, err)
	}

	return 0
}

//export goavFileSeek
func goavFileSeek(p unsafe.Pointer, offset C.int64_t, whence C.int) C.int64_t {
	pos, err := getPinnedFile(p).Seek(int64(offset), int(whence))
	if err != nil {
		return C.int64_t(returnPinnedFileError(p, err))
	}

	return C.int64_t(pos)
}

type _ioContext = avformat.IOContext

type ioContext struct {
	*_ioContext
}

func allocAvioContext(f interface{}, writable bool) *avformat.IOContext {
	var read, write, seek *[0]byte
	if _, ok := f.(io.Reader); ok {
		read = (*[0]byte)(C.goavFileReadPacket)
	}

	if _, ok := f.(io.Writer); ok {
		write = (*[0]byte)(C.goavFileWritePacket)
	}

	if _, ok := f.(io.Seeker); ok {
		seek = (*[0]byte)(C.goavFileSeek)
	}

	if read == nil && write == nil {
		panic("f must implement at least one of io.Reader or io.Writer")
	}

	var p pinType
	for {
		p = randPin()
		if _, ok := pinnedFiles[p]; !ok {
			break
		}
	}

	var writeFlag int32
	if writable {
		writeFlag = 1
	}

	ctx := avformat.NewIOContext((*byte)(avutil.Malloc(1<<12)), 1<<12, writeFlag, pinptr(p), read, write, seek)
	if ctx == nil {
		panic(ErrNoMem)
	}

	pinnedFiles[p] = &pinnedFile{f: f}

	return ctx
}

func newIOContext(f interface{}, writable bool) *ioContext {
	ret := &ioContext{
		_ioContext: allocAvioContext(f, writable),
	}

	runtime.SetFinalizer(ret, func(ctx *ioContext) {
		delete(pinnedFiles, pin(ctx._ioContext.Opaque))
		avformat.FreeIOContext(&ctx._ioContext)
	})

	return ret
}
