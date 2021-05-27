package av

import (
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/ssttevee/go-av/avfilter"
	"github.com/ssttevee/go-av/avutil"
)

type FilterNotFoundError string

func (e FilterNotFoundError) Error() string {
	return "filter not found: " + string(e)
}

type _filter = avfilter.Filter

type Filter struct {
	*_filter
}

func FindFilterByName(name string) (*Filter, error) {
	filter := avfilter.GetByName(name)
	if filter == nil {
		return nil, errors.WithStack(FilterNotFoundError(name))
	}

	return &Filter{_filter: filter}, nil
}

type FilterInOut struct {
	Name          string
	FilterContext *FilterContext
	PadIndex      int32
}

type _filterGraph = avfilter.Graph

type FilterGraph struct {
	*_filterGraph

	initOnce sync.Once
	initErr  error
}

func NewFilterGraph() (*FilterGraph, error) {
	graph := avfilter.NewGraph()
	if graph == nil {
		panic(avutil.ErrNoMem)
	}

	ret := &FilterGraph{_filterGraph: graph}

	runtime.SetFinalizer(ret, func(graph *FilterGraph) {
		// heap pointer may not be passed to cgo, so use a stack pointer instead :D
		filterGraph := (*avfilter.Graph)(graph._filterGraph)
		avfilter.FreeGraph(&filterGraph)
		graph._filterGraph = filterGraph
	})

	return ret, nil
}

func (g *FilterGraph) wrapFilterIO(io *avfilter.InOut) *FilterInOut {
	return &FilterInOut{
		Name:          io.Name.String(),
		FilterContext: &FilterContext{g: g, _filterContext: io.FilterCtx},
		PadIndex:      io.PadIdx,
	}
}

func (g *FilterGraph) Parse(desc string) (inputs, outputs []*FilterInOut, _ error) {
	var cinputs, coutputs *avfilter.InOut
	if err := averror(avfilter.ParseGraph(g._filterGraph, desc, &cinputs, &coutputs)); err != nil {
		return nil, nil, err
	}

	defer avfilter.FreeInOut(&cinputs)
	defer avfilter.FreeInOut(&coutputs)

	for current := cinputs; current != nil; current = current.Next {
		inputs = append(inputs, g.wrapFilterIO(current))
	}

	for current := coutputs; current != nil; current = current.Next {
		outputs = append(outputs, g.wrapFilterIO(current))
	}

	return inputs, outputs, nil
}

func (g *FilterGraph) init() error {
	g.initOnce.Do(func() {
		g.initErr = averror(avfilter.ConfigGraph(g._filterGraph, nil))
	})

	return g.initErr
}

func (g *FilterGraph) filters() []*avfilter.Context {
	return *(*[]*avfilter.Context)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(g._filterGraph.Filters)),
		Len:  int(g._filterGraph.NbFilters),
		Cap:  int(g._filterGraph.NbFilters),
	}))
}

func (g *FilterGraph) SetHWDeviceContext(ctx *HWDeviceContext) {
	for _, filter := range g.filters() {
		filter.HwDeviceCtx = ctx.ref()
	}
}

func (g *FilterGraph) newFilter(filter *Filter, name, args string) (*avfilter.Context, error) {
	var ctx *avfilter.Context
	if err := averror(avfilter.CreateFilterGraph(&ctx, filter._filter, name, args, nil, g._filterGraph)); err != nil {
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
		g:              g,
		_filterContext: ctx,
	}, nil
}

func (g *FilterGraph) NewFilterByName(filterName, name, args string) (*FilterContext, error) {
	filter, err := FindFilterByName(filterName)
	if err != nil {
		return nil, err
	}

	return g.NewFilter(filter, name, args)
}

func linkFilters(src *FilterContext, srcPadIndex int32, dst *FilterContext, dstPadIndex int32) error {
	return averror(avfilter.Link(src._filterContext, uint32(srcPadIndex), dst._filterContext, uint32(dstPadIndex)))
}

type _filterContext = avfilter.Context

type FilterContext struct {
	*_filterContext
	g *FilterGraph
}

func (ctx *FilterContext) LinkTo(padIndex int32, dst *FilterContext, dstPadIndex int32) error {
	return linkFilters(ctx, padIndex, dst, dstPadIndex)
}

func (ctx *FilterContext) Name() string {
	return ctx._filterContext.Name.String()
}

func (ctx *FilterContext) LinkFrom(padIndex int32, src *FilterContext, srcPadIndex int32) error {
	return linkFilters(src, srcPadIndex, ctx, padIndex)
}

type BufferSource FilterContext

func (g *FilterGraph) NewBufferSource(name string, decoder *DecoderContext) (*BufferSource, error) {
	filter, err := FindFilterByName("buffer")
	if err != nil {
		return nil, err
	}

	par := avfilter.NewBufferSourceParameters()
	if par == nil {
		panic(avutil.ErrNoMem)
	}

	defer avutil.Free(unsafe.Pointer(par))

	if decoder._codecContext.HwFramesCtx != nil {
		par.HwFramesCtx = avutil.RefBuffer(decoder._codecContext.HwFramesCtx)
		if par.HwFramesCtx == nil {
			panic(avutil.ErrNoMem)
		}
	}

	par.Format = decoder.PixFmt

	ctx, err := g.newFilter(filter, name, decoder.BufferSourceArgs())
	if err != nil {
		return nil, err
	}

	if decoder._codecContext.HwFramesCtx != nil {
		ctx.HwDeviceCtx = avutil.RefBuffer(decoder._codecContext.HwFramesCtx)
		if ctx.HwDeviceCtx == nil {
			panic(avutil.ErrNoMem)
		}
	}

	if err := averror(avfilter.SetBufferSourceParameters(ctx, par)); err != nil {
		return nil, err
	}

	return &BufferSource{
		g:              g,
		_filterContext: ctx,
	}, nil
}

func (src *BufferSource) WriteFrame(frame *Frame) error {
	if err := src.g.init(); err != nil {
		return err
	}

	defer runtime.KeepAlive(frame)

	return averror(avfilter.WriteBufferSourceFrame(src._filterContext, frame._frame))
}

func (src *BufferSource) LinkTo(dst *FilterContext, dstPadIndex int32) error {
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
		g:              g,
		_filterContext: ctx,
	}, nil
}

func (sink *BufferSink) ReadFrameReuse(frame *Frame) error {
	if err := sink.g.init(); err != nil {
		return err
	}

	return averror(avfilter.GetBufferSinkFrame(sink._filterContext, frame.prepare()))
}

func (sink *BufferSink) ReadFrame() (*Frame, error) {
	frame := NewFrame()
	if err := sink.ReadFrameReuse(frame); errors.Is(err, avutil.ErrAgain) {
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

func (sink *BufferSink) LinkFrom(src *FilterContext, srcPadIndex int32) error {
	return linkFilters(src, srcPadIndex, (*FilterContext)(sink), 0)
}
