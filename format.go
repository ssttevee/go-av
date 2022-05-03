package av

// struct AVFormatContext;
// struct AVIOContext;
// struct AVDictionary;
//
// extern int goavIOOpen(struct AVFormatContext *s, struct AVIOContext **pb, char *url, int flags, struct AVDictionary **options);
// extern int goavIOClose(struct AVFormatContext *s, struct AVIOContext *pb);
import "C"
import (
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"reflect"
	"runtime"
	"runtime/cgo"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/ssttevee/go-av/avcodec"
	"github.com/ssttevee/go-av/avformat"
	"github.com/ssttevee/go-av/avutil"
	"github.com/ssttevee/go-av/internal/common"
)

type Opener interface {
	Open(url string, flags int) (io.Closer, error)
}

const FileOpener = fileOpener(0)

type fileOpener int

func (fileOpener) Open(inputURL string, flags int) (io.Closer, error) {
	u, err := url.Parse(inputURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	const scheme = "file"
	if u.Scheme != "" && u.Scheme != scheme {
		return nil, errors.Errorf("expected %q scheme, but got %q", scheme, u.Scheme)
	}

	var filepath string
	if u.Opaque != "" {
		filepath = u.Opaque
	} else {
		filepath = u.Path
	}

	dir, _ := path.Split(filepath)
	if dir != "" {
		if err := os.MkdirAll(dir, 0666); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	f, err := os.OpenFile(filepath, flags, 0666)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return f, nil
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error {
	return nil
}

const NullOpener = nullOpener(0)

type nullOpener int

func (nullOpener) Open(inputURL string, flags int) (io.Closer, error) {
	if flags&os.O_WRONLY == 0 {
		return nil, errors.Errorf("unsupported flags: %b", flags)
	}

	return nopCloser{
		Writer: ioutil.Discard,
	}, nil
}

type pinnedFormatContextData struct {
	opener Opener
	err    error
}

func unwrapPinnedFormatContextDataEntries(p unsafe.Pointer) *pinnedFormatContextData {
	return cgo.Handle(p).Value().(*pinnedFormatContextData)
}

func returnPinnedFormatContextDataError(p unsafe.Pointer, err error) C.int {
	unwrapPinnedFormatContextDataEntries(p).err = err
	return C.int(common.FormatError)
}

//export goavIOOpen
func goavIOOpen(s *C.struct_AVFormatContext, pb **C.struct_AVIOContext, url *C.char, flags C.int, options **C.struct_AVDictionary) C.int {
	var goflags int
	var writable bool
	if flags&avformat.IOFlagWrite != 0 {
		goflags = os.O_WRONLY | os.O_CREATE
		writable = true
	} else if flags&avformat.IOFlagRead != 0 {
		goflags = os.O_RDONLY
	}

	opaque := (*avformat.Context)(unsafe.Pointer(s)).Opaque
	defer runtime.KeepAlive(opaque)

	f, err := unwrapPinnedFormatContextDataEntries(opaque).opener.Open(C.GoString(url), goflags)
	if err != nil {
		return returnPinnedFormatContextDataError(opaque, err)
	}

	*pb = (*C.struct_AVIOContext)(unsafe.Pointer(allocAvioContext(f, writable)))

	return 0
}

//export goavIOClose
func goavIOClose(s *C.struct_AVFormatContext, pb *C.struct_AVIOContext) C.int {
	opaque := (*avformat.IOContext)(unsafe.Pointer(pb)).Opaque
	defer runtime.KeepAlive(opaque)

	if err := unwrapPinnedFile(opaque).f.(io.Closer).Close(); err != nil {
		return returnPinnedFormatContextDataError((*avformat.Context)(unsafe.Pointer(s)).Opaque, err)
	}

	cgo.Handle(opaque).Delete()

	return 0
}

type _formatContext = avformat.Context

type formatContext struct {
	*_formatContext

	pinnedDataOnce sync.Once
}

func (ctx *formatContext) pinnedData() *pinnedFormatContextData {
	ctx.pinnedDataOnce.Do(func() {
		ctx.Opaque = unsafe.Pointer(cgo.NewHandle(&pinnedFormatContextData{}))
	})

	return unwrapPinnedFormatContextDataEntries(ctx.Opaque)
}

func (ctx *formatContext) finalizePinnedData() {
	if ctx.Opaque == nil {
		return
	}

	cgo.Handle(ctx.Opaque).Delete()
}

func (ctx *formatContext) FindBestStream(mediaType avutil.MediaType) (int, *Codec, error) {
	var codec *avcodec.Codec
	streamIndex, err := avreturn(avformat.FindBestStream(ctx._formatContext, mediaType, -1, -1, &codec, 0))
	if errors.Is(err, avutil.ErrStreamNotFound) {
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
