// +godefs map struct_AVFormatContext int64

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

	"github.com/ssttevee/go-av/avcodec"
	"github.com/ssttevee/go-av/avformat"
	"github.com/ssttevee/go-av/avutil"
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

	*pb = (*C.AVIOContext)(unsafe.Pointer(allocAvioContext(f, writable)))

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

type _formatContext = avformat.Context

type formatContext struct {
	*_formatContext

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

		ctx.Opaque = pinptr(p)
	})

	return pinnedFormatContextDataEntries[pin(ctx.Opaque)]
}

func (ctx *formatContext) finalizePinnedData() {
	if ctx.Opaque == nil {
		return
	}

	delete(pinnedFormatContextDataEntries, pin(ctx.Opaque))
}

func (ctx *formatContext) FindBestStream(mediaType avformat.MediaType) (int, *Codec, error) {
	var codec *avcodec.Codec
	streamIndex, err := avreturn(avformat.FindBestStream(ctx._formatContext, mediaType, -1, -1, &codec, 0))
	if err == Error(C.AVERROR_STREAM_NOT_FOUND) {
		return -1, nil, nil
	} else if err != nil {
		return 0, nil, err
	}

	return streamIndex, &Codec{_codec: codec}, nil
}

func (ctx *formatContext) GuessFramerate(stream *Stream) avutil.Rational {
	return avformat.GuessFrameRate(ctx._formatContext, stream._stream, nil)
}

func (ctx *formatContext) SetOption(name string, value interface{}) error {
	return setOption(ctx.PrivData, name, value, 0)
}

func (ctx *formatContext) GetOption(name string) (interface{}, error) {
	return getOption(ctx.PrivData, name, 0)
}

func (ctx *formatContext) streams() []*avformat.Stream {
	return *(*[]*avformat.Stream)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(unsafe.Pointer(ctx._formatContext.Streams)), Len: int(ctx.NbStreams), Cap: int(ctx.NbStreams)}))
}

func (ctx *formatContext) Streams() []*Stream {
	streams := ctx.streams()
	ret := make([]*Stream, len(streams))
	for i, stream := range streams {
		ret[i] = &Stream{
			_stream:   stream,
			formatCtx: ctx._formatContext,
		}
	}

	return ret
}

func (ctx *formatContext) Stream(i int) *Stream {
	return &Stream{
		_stream:   ctx.streams()[i],
		formatCtx: ctx._formatContext,
	}
}

func (ctx *formatContext) Url() string {
	return ctx._formatContext.Url.String()
}

func (ctx *formatContext) SetUrl(url string) {
	if ctx._formatContext.Url != nil {
		avutil.Free(unsafe.Pointer(ctx._formatContext.Url))
	}

	ctx._formatContext.Url = avutil.DupeString(url)
}

func (ctx *formatContext) SetOpener(opener Opener) {
	ctx.pinnedData().opener = opener
	ctx.IoOpen = (*[0]byte)(C.goavIOOpen)
	ctx.IoClose = (*[0]byte)(C.goavIOClose)
}
