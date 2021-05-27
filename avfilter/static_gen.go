// Code generated by robots; DO NOT EDIT.
// +build !av_dynamic

package avfilter

import (
	avutil "github.com/ssttevee/go-av/avutil"
	"runtime"
	"unsafe"
)

/*
#include <libavfilter/buffersrc.h>
#include <libavfilter/buffersink.h>
*/
import "C"

func ConfigGraph(p0 *Graph, p1 unsafe.Pointer) int32 {
	defer runtime.KeepAlive(p0)
	ret := C.avfilter_graph_config((*C.struct_AVFilterGraph)(unsafe.Pointer(p0)), p1)
	return *(*int32)(unsafe.Pointer(&ret))
}
func CreateFilterGraph(p0 **Context, p1 *Filter, p2 string, p3 string, p4 unsafe.Pointer, p5 *Graph) int32 {
	defer runtime.KeepAlive(p0)
	defer runtime.KeepAlive(p1)
	var s2 *C.char
	if p2 != "" {
		s2 = C.CString(p2)
		defer C.free(unsafe.Pointer(s2))
	}
	var s3 *C.char
	if p3 != "" {
		s3 = C.CString(p3)
		defer C.free(unsafe.Pointer(s3))
	}
	defer runtime.KeepAlive(p5)
	ret := C.avfilter_graph_create_filter((**C.struct_AVFilterContext)(unsafe.Pointer(p0)), (*C.struct_AVFilter)(unsafe.Pointer(p1)), s2, s3, p4, (*C.struct_AVFilterGraph)(unsafe.Pointer(p5)))
	return *(*int32)(unsafe.Pointer(&ret))
}
func FreeGraph(p0 **Graph) {
	defer runtime.KeepAlive(p0)
	C.avfilter_graph_free((**C.struct_AVFilterGraph)(unsafe.Pointer(p0)))
}
func FreeInOut(p0 **InOut) {
	defer runtime.KeepAlive(p0)
	C.avfilter_inout_free((**C.struct_AVFilterInOut)(unsafe.Pointer(p0)))
}
func GetBufferSinkFrame(p0 *Context, p1 *avutil.Frame) int32 {
	defer runtime.KeepAlive(p0)
	defer runtime.KeepAlive(p1)
	ret := C.av_buffersink_get_frame((*C.struct_AVFilterContext)(unsafe.Pointer(p0)), (*C.struct_AVFrame)(unsafe.Pointer(p1)))
	return *(*int32)(unsafe.Pointer(&ret))
}
func GetByName(p0 string) *Filter {
	var s0 *C.char
	if p0 != "" {
		s0 = C.CString(p0)
		defer C.free(unsafe.Pointer(s0))
	}
	return (*Filter)(unsafe.Pointer(C.avfilter_get_by_name(s0)))
}
func Link(p0 *Context, p1 uint32, p2 *Context, p3 uint32) int32 {
	defer runtime.KeepAlive(p0)
	defer runtime.KeepAlive(p1)
	defer runtime.KeepAlive(p2)
	defer runtime.KeepAlive(p3)
	ret := C.avfilter_link((*C.struct_AVFilterContext)(unsafe.Pointer(p0)), *(*C.uint)(unsafe.Pointer(&p1)), (*C.struct_AVFilterContext)(unsafe.Pointer(p2)), *(*C.uint)(unsafe.Pointer(&p3)))
	return *(*int32)(unsafe.Pointer(&ret))
}
func NewBufferSourceParameters() *BufferSourceParameters {
	return (*BufferSourceParameters)(unsafe.Pointer(C.av_buffersrc_parameters_alloc()))
}
func NewGraph() *Graph {
	return (*Graph)(unsafe.Pointer(C.avfilter_graph_alloc()))
}
func ParseGraph(p0 *Graph, p1 string, p2 **InOut, p3 **InOut) int32 {
	defer runtime.KeepAlive(p0)
	var s1 *C.char
	if p1 != "" {
		s1 = C.CString(p1)
		defer C.free(unsafe.Pointer(s1))
	}
	defer runtime.KeepAlive(p2)
	defer runtime.KeepAlive(p3)
	ret := C.avfilter_graph_parse2((*C.struct_AVFilterGraph)(unsafe.Pointer(p0)), s1, (**C.struct_AVFilterInOut)(unsafe.Pointer(p2)), (**C.struct_AVFilterInOut)(unsafe.Pointer(p3)))
	return *(*int32)(unsafe.Pointer(&ret))
}
func SetBufferSourceParameters(p0 *Context, p1 *BufferSourceParameters) int32 {
	defer runtime.KeepAlive(p0)
	defer runtime.KeepAlive(p1)
	ret := C.av_buffersrc_parameters_set((*C.struct_AVFilterContext)(unsafe.Pointer(p0)), (*C.struct_AVBufferSrcParameters)(unsafe.Pointer(p1)))
	return *(*int32)(unsafe.Pointer(&ret))
}
func WriteBufferSourceFrame(p0 *Context, p1 *avutil.Frame) int32 {
	defer runtime.KeepAlive(p0)
	defer runtime.KeepAlive(p1)
	ret := C.av_buffersrc_write_frame((*C.struct_AVFilterContext)(unsafe.Pointer(p0)), (*C.struct_AVFrame)(unsafe.Pointer(p1)))
	return *(*int32)(unsafe.Pointer(&ret))
}
