package av

// #include <libavfilter/avfilter.h>
// #include <libavfilter/buffersrc.h>
// #include <libavfilter/buffersink.h>
// #include <libavutil/hwcontext.h>
//
// typedef struct HWDownloadContext {
//     const AVClass *class;
//
//     AVBufferRef       *hwframes_ref;
//     AVHWFramesContext *hwframes;
// } HWDownloadContext;
import "C"
import (
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
)

type FilterNotFoundError string

func (e FilterNotFoundError) Error() string {
	return "filter not found: " + string(e)
}

type Filter struct {
	filter *C.AVFilter
}

func findFilterByName(name string) *C.AVFilter {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return C.avfilter_get_by_name(cname)
}

func FindFilterByName(name string) (*Filter, error) {
	filter := findFilterByName(name)
	if filter == nil {
		return nil, errors.WithStack(FilterNotFoundError(name))
	}

	return &Filter{filter: filter}, nil
}

type FilterInOut struct {
	Name          string
	FilterContext *FilterContext
	PadIndex      int
}

type FilterGraph struct {
	graph *C.AVFilterGraph

	initOnce sync.Once
	initErr  error
}

func NewFilterGraph() (*FilterGraph, error) {
	graph := C.avfilter_graph_alloc()
	if graph == nil {
		panic(ErrNoMem)
	}

	ret := &FilterGraph{graph: graph}

	runtime.SetFinalizer(ret, func(graph *FilterGraph) {
		C.avfilter_graph_free(&graph.graph)
	})

	return ret, nil
}

func (g *FilterGraph) wrapFilterIO(io *C.AVFilterInOut) *FilterInOut {
	return &FilterInOut{
		Name:          C.GoString(io.name),
		FilterContext: &FilterContext{g: g, ctx: io.filter_ctx},
		PadIndex:      int(io.pad_idx),
	}
}

func (g *FilterGraph) Parse(desc string) (inputs, outputs []*FilterInOut, _ error) {
	cdesc := C.CString(desc)
	defer C.free(unsafe.Pointer(cdesc))

	var cinputs, coutputs *C.AVFilterInOut
	if err := averror(C.avfilter_graph_parse2(g.graph, cdesc, &cinputs, &coutputs)); err != nil {
		return nil, nil, err
	}

	defer C.avfilter_inout_free(&cinputs)
	defer C.avfilter_inout_free(&coutputs)

	for current := cinputs; current != nil; current = current.next {
		inputs = append(inputs, g.wrapFilterIO(current))
	}

	for current := coutputs; current != nil; current = current.next {
		outputs = append(outputs, g.wrapFilterIO(current))
	}

	return inputs, outputs, nil
}

func (g *FilterGraph) init() error {
	g.initOnce.Do(func() {
		g.initErr = averror(C.avfilter_graph_config(g.graph, nil))
	})

	return g.initErr
}

func (g *FilterGraph) filters() []*C.AVFilterContext {
	return *(*[]*C.AVFilterContext)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(g.graph.filters)),
		Len:  int(g.graph.nb_filters),
		Cap:  int(g.graph.nb_filters),
	}))
}

func (g *FilterGraph) SetHWDeviceContext(ctx *HWDeviceContext) {
	for _, filter := range g.filters() {
		filter.hw_device_ctx = ctx.ref()
	}
}

func (g *FilterGraph) newFilter(filter *Filter, name, args string) (*C.AVFilterContext, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cargs := C.CString(args)
	defer C.free(unsafe.Pointer(cargs))

	var ctx *C.AVFilterContext
	if err := averror(C.avfilter_graph_create_filter(&ctx, filter.filter, cname, cargs, nil, g.graph)); err != nil {
		return nil, err
	}

	return ctx, nil
}

func (g *FilterGraph) NewFilter(filter *Filter, name, args string) (*FilterContext, error) {
	ctx, err := g.newFilter(filter, name, args)
	if err != nil {
		return nil, err
	}

	return &FilterContext{
		g:   g,
		ctx: ctx,
	}, nil
}

func (g *FilterGraph) NewFilterByName(filterName, name, args string) (*FilterContext, error) {
	filter, err := FindFilterByName(filterName)
	if err != nil {
		return nil, err
	}

	return g.NewFilter(filter, name, args)
}

func linkFilters(src *FilterContext, srcPadIndex int, dst *FilterContext, dstPadIndex int) error {
	return averror(C.avfilter_link(src.ctx, C.uint(srcPadIndex), dst.ctx, C.uint(dstPadIndex)))
}

type FilterContext struct {
	g   *FilterGraph
	ctx *C.AVFilterContext
}

func (ctx *FilterContext) LinkTo(padIndex int, dst *FilterContext, dstPadIndex int) error {
	return linkFilters(ctx, padIndex, dst, dstPadIndex)
}

func (ctx *FilterContext) Name() string {
	return C.GoString(ctx.ctx.name)
}

func (ctx *FilterContext) HWDownloadContext() *C.HWDownloadContext {
	return (*C.HWDownloadContext)(ctx.ctx.priv)
}

func (ctx *FilterContext) LinkFrom(padIndex int, src *FilterContext, srcPadIndex int) error {
	return linkFilters(src, srcPadIndex, ctx, padIndex)
}

type BufferSource FilterContext

func (g *FilterGraph) NewBufferSource(name string, decoder *DecoderContext) (*BufferSource, error) {
	filter, err := FindFilterByName("buffer")
	if err != nil {
		return nil, err
	}

	par := C.av_buffersrc_parameters_alloc()
	if par == nil {
		panic(ErrNoMem)
	}

	defer C.av_free(unsafe.Pointer(par))

	if decoder.ctx.hw_frames_ctx != nil {
		par.hw_frames_ctx = C.av_buffer_ref(decoder.ctx.hw_frames_ctx)
		if par.hw_frames_ctx == nil {
			panic(ErrNoMem)
		}
	}

	par.format = C.int(decoder.PixelFormat().ctype())

	ctx, err := g.newFilter(filter, name, decoder.BufferSourceArgs())
	if err != nil {
		return nil, err
	}

	if decoder.ctx.hw_device_ctx != nil {
		ctx.hw_device_ctx = C.av_buffer_ref(decoder.ctx.hw_device_ctx)
		if ctx.hw_device_ctx == nil {
			panic(ErrNoMem)
		}
	}

	if err := averror(C.av_buffersrc_parameters_set(ctx, par)); err != nil {
		return nil, err
	}

	return &BufferSource{
		g:   g,
		ctx: ctx,
	}, nil
}

func (src *BufferSource) WriteFrame(frame *Frame) error {
	if err := src.g.init(); err != nil {
		return err
	}

	defer runtime.KeepAlive(frame)

	return averror(C.av_buffersrc_write_frame(src.ctx, frame.frame))
}

func (src *BufferSource) LinkTo(dst *FilterContext, dstPadIndex int) error {
	return linkFilters((*FilterContext)(src), 0, dst, dstPadIndex)
}

type BufferSink FilterContext

func (g *FilterGraph) NewBufferSink(name string) (*BufferSink, error) {
	filter, err := FindFilterByName("buffersink")
	if err != nil {
		return nil, err
	}

	ctx, err := g.newFilter(filter, name, "")
	if err != nil {
		return nil, err
	}

	return &BufferSink{
		g:   g,
		ctx: ctx,
	}, nil
}

func (sink *BufferSink) ReadFrameReuse(frame *Frame) error {
	if err := sink.g.init(); err != nil {
		return err
	}

	return averror(C.av_buffersink_get_frame(sink.ctx, frame.prepare()))
}

func (sink *BufferSink) ReadFrame() (*Frame, error) {
	frame := NewFrame()
	if err := sink.ReadFrameReuse(frame); errors.Is(err, ErrAgain) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return frame, nil
}

func (sink *BufferSink) ReadFrames() ([]*Frame, error) {
	var frames []*Frame
	for {
		frame, err := sink.ReadFrame()
		if err != nil {
			return nil, err
		} else if frame == nil {
			break
		}

		frames = append(frames, frame)
	}

	return frames, nil
}

func (sink *BufferSink) LinkFrom(src *FilterContext, srcPadIndex int) error {
	return linkFilters(src, srcPadIndex, (*FilterContext)(sink), 0)
}
