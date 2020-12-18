package av

// #include <libavformat/avformat.h>
//
// extern int goavIOOpen(struct AVFormatContext *s, AVIOContext **pb, char *url, int flags, AVDictionary **options);
// extern int goavIOClose(struct AVFormatContext *s, AVIOContext *pb);
import "C"
import (
	"io"
	"os"
	"reflect"
	"sync"
	"unsafe"
)

type Opener interface {
	Open(url string, flags int) (io.Closer, error)
}

type pinnedFormatContextData struct {
	opener Opener
	err    error
}

var pinnedFormatContextDataEntries = map[pinType]*pinnedFormatContextData{}

func returnPinnedFormatContextDataError(p unsafe.Pointer, err error) C.int {
	pinnedFormatContextDataEntries[pin(p)].err = err
	return C.int(errInternalFormatError)
}

//export goavIOOpen
func goavIOOpen(s *C.struct_AVFormatContext, pb **C.AVIOContext, url *C.char, flags C.int, options **C.AVDictionary) C.int {
	data, ok := pinnedFormatContextDataEntries[pin(s.opaque)]
	if !ok {
		panic("pinned data not found")
	}

	var goflags int
	var writable bool
	if flags&C.AVIO_FLAG_WRITE != 0 {
		goflags = os.O_WRONLY | os.O_CREATE
		writable = true
	} else if flags&C.AVIO_FLAG_READ != 0 {
		goflags = os.O_RDONLY
	}

	f, err := data.opener.Open(C.GoString(url), goflags)
	if err != nil {
		return returnPinnedFormatContextDataError(s.opaque, err)
	}

	*pb = allocAvioContext(f, writable)

	return 0
}

//export goavIOClose
func goavIOClose(s *C.struct_AVFormatContext, pb *C.AVIOContext) C.int {
	if err := pinnedFiles[pin(pb.opaque)].f.(io.Closer).Close(); err != nil {
		return returnPinnedFormatContextDataError(s.opaque, err)
	}

	delete(pinnedFiles, pin(pb.opaque))

	return 0
}

type formatContext struct {
	ctx *C.AVFormatContext

	pinnedDataOnce sync.Once
}

func (ctx *formatContext) pinnedData() *pinnedFormatContextData {
	ctx.pinnedDataOnce.Do(func() {
		var p pinType
		for {
			p = randPin()
			if _, ok := pinnedFormatContextDataEntries[p]; !ok {
				break
			}
		}

		pinnedFormatContextDataEntries[p] = &pinnedFormatContextData{}

		ctx.ctx.opaque = pinptr(p)
	})

	return pinnedFormatContextDataEntries[pin(ctx.ctx.opaque)]
}

func (ctx *formatContext) finalizedPinnedData() {
	if ctx.ctx.opaque == nil {
		return
	}

	delete(pinnedFormatContextDataEntries, pin(ctx.ctx.opaque))
}

func (ctx *formatContext) FindBestStream(mediaType MediaType) (int, *Codec, error) {
	var codec *C.AVCodec
	streamIndex, err := avreturn(C.av_find_best_stream(ctx.ctx, mediaType.ctype(), -1, -1, &codec, 0))
	if err == Error(C.AVERROR_STREAM_NOT_FOUND) {
		return -1, nil, nil
	} else if err != nil {
		return 0, nil, err
	}

	return streamIndex, &Codec{codec: codec}, nil
}

func (ctx *formatContext) GuessFramerate(stream *Stream) Rational {
	return rational(C.av_guess_frame_rate(ctx.ctx, stream.stream, nil))
}

func (ctx *formatContext) SetOption(name string, value interface{}) error {
	return setOption(ctx.ctx.priv_data, name, value, 0)
}

func (ctx *formatContext) GetOption(name string) (interface{}, error) {
	return getOption(ctx.ctx.priv_data, name, 0)
}

func (ctx *formatContext) NumStream() int {
	return int(ctx.ctx.nb_streams)
}

func (ctx *formatContext) Stream(i int) *Stream {
	return &Stream{
		stream:    (*(*[]*C.AVStream)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(unsafe.Pointer(ctx.ctx.streams)), Len: int(ctx.ctx.nb_streams), Cap: int(ctx.ctx.nb_streams)})))[i],
		formatCtx: ctx.ctx,
	}
}

func (ctx *formatContext) URL() string {
	return C.GoString(ctx.ctx.url)
}

func (ctx *formatContext) SetURL(url string) {
	if ctx.ctx.url != nil {
		C.av_free(unsafe.Pointer(ctx.ctx.url))
	}

	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	ctx.ctx.url = C.av_strdup(curl)
}

func (ctx *formatContext) SetOpener(opener Opener) {
	ctx.pinnedData().opener = opener
	ctx.ctx.io_open = (*[0]byte)(C.goavIOOpen)
	ctx.ctx.io_close = (*[0]byte)(C.goavIOClose)
}
