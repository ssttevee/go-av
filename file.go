package av

// #include <stdint.h>
//
// extern int goavFileReadPacket(void *opaque, uint8_t *buf, int buf_size);
// extern int goavFileWritePacket(void *opaque, uint8_t *buf, int buf_size);
// extern int64_t goavFileSeek(void *opaque, int64_t offset, int whence);
import "C"
import (
	"io"
	"reflect"
	"runtime"
	"runtime/cgo"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/ssttevee/go-av/avformat"
	"github.com/ssttevee/go-av/avutil"
	"github.com/ssttevee/go-av/internal/common"
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

func unwrapPinnedFile(p unsafe.Pointer) *pinnedFile {
	return cgo.Handle(p).Value().(*pinnedFile)
}

func getPinnedFile(p unsafe.Pointer) io.ReadWriteSeeker {
	return unwrapPinnedFile(p)
}

func returnPinnedFileError(p unsafe.Pointer, err error) C.int {
	unwrapPinnedFile(p).err = err
	return C.int(common.IOError)
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
	if errors.Is(err, io.EOF) {
		if n <= 0 {
			return C.int(avutil.ErrEOF)
		}
	} else if err != nil {
		return returnPinnedFileError(p, errors.WithStack(err))
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
	var read, write, seek unsafe.Pointer
	if _, ok := f.(io.Reader); ok {
		read = C.goavFileReadPacket
	}

	if _, ok := f.(io.Writer); ok {
		write = C.goavFileWritePacket
	}

	if _, ok := f.(io.Seeker); ok {
		seek = C.goavFileSeek
	}

	if read == nil && write == nil {
		panic("f must implement at least one of io.Reader or io.Writer")
	}

	var writeFlag int32
	if writable {
		writeFlag = 1
	}

	h := cgo.NewHandle(&pinnedFile{f: f})
	ctx := avformat.NewIOContext((*byte)(avutil.Malloc(1<<12)), 1<<12, writeFlag, unsafe.Pointer(h), read, write, seek)
	if ctx == nil {
		h.Delete()
		panic(avutil.ErrNoMem)
	}

	return ctx
}

func newIOContext(f interface{}, writable bool) *ioContext {
	ret := &ioContext{
		_ioContext: allocAvioContext(f, writable),
	}

	runtime.SetFinalizer(ret, func(ctx *ioContext) {
		cgo.Handle(ctx._ioContext.Opaque).Delete()
		// heap pointer may not be passed to cgo, so use a stack pointer instead :D
		ioContext := (*avformat.IOContext)(ctx._ioContext)
		avformat.FreeIOContext(&ioContext)
		ctx._ioContext = ioContext
	})

	return ret
}
