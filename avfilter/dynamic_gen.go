// Code generated by robots; DO NOT EDIT.
//go:build av_dynamic
// +build av_dynamic

package avfilter

import (
	errors "github.com/pkg/errors"
	avutil "github.com/ssttevee/go-av/avutil"
	"runtime"
	"sync"
	"unsafe"
)

/*
#cgo LDFLAGS: -ldl

#include <stdlib.h>
#include <stdint.h>
#include <dlfcn.h>

struct AVBufferSrcParameters;
struct AVFilter;
struct AVFilterContext;
struct AVFilterGraph;
struct AVFilterInOut;
struct AVFrame;

static void *handle = 0;

static int (*_avfilter_graph_config)(struct AVFilterGraph*, void*);

int dyn_avfilter_graph_config(struct AVFilterGraph* p0, void* p1) {
    return _avfilter_graph_config(p0, p1);
};

static int (*_avfilter_graph_create_filter)(struct AVFilterContext**, struct AVFilter*, char*, char*, void*, struct AVFilterGraph*);

int dyn_avfilter_graph_create_filter(struct AVFilterContext** p0, struct AVFilter* p1, char* p2, char* p3, void* p4, struct AVFilterGraph* p5) {
    return _avfilter_graph_create_filter(p0, p1, p2, p3, p4, p5);
};

static void (*_avfilter_graph_free)(struct AVFilterGraph**);

void dyn_avfilter_graph_free(struct AVFilterGraph** p0) {
    _avfilter_graph_free(p0);
};

static void (*_avfilter_inout_free)(struct AVFilterInOut**);

void dyn_avfilter_inout_free(struct AVFilterInOut** p0) {
    _avfilter_inout_free(p0);
};

static int (*_av_buffersink_get_frame)(struct AVFilterContext*, struct AVFrame*);

int dyn_av_buffersink_get_frame(struct AVFilterContext* p0, struct AVFrame* p1) {
    return _av_buffersink_get_frame(p0, p1);
};

static struct AVFilter* (*_avfilter_get_by_name)(char*);

struct AVFilter* dyn_avfilter_get_by_name(char* p0) {
    return _avfilter_get_by_name(p0);
};

static int (*_avfilter_link)(struct AVFilterContext*, uint, struct AVFilterContext*, uint);

int dyn_avfilter_link(struct AVFilterContext* p0, uint p1, struct AVFilterContext* p2, uint p3) {
    return _avfilter_link(p0, p1, p2, p3);
};

static struct AVBufferSrcParameters* (*_av_buffersrc_parameters_alloc)();

struct AVBufferSrcParameters* dyn_av_buffersrc_parameters_alloc() {
    return _av_buffersrc_parameters_alloc();
};

static struct AVFilterGraph* (*_avfilter_graph_alloc)();

struct AVFilterGraph* dyn_avfilter_graph_alloc() {
    return _avfilter_graph_alloc();
};

static int (*_avfilter_graph_parse2)(struct AVFilterGraph*, char*, struct AVFilterInOut**, struct AVFilterInOut**);

int dyn_avfilter_graph_parse2(struct AVFilterGraph* p0, char* p1, struct AVFilterInOut** p2, struct AVFilterInOut** p3) {
    return _avfilter_graph_parse2(p0, p1, p2, p3);
};

static int (*_av_buffersrc_parameters_set)(struct AVFilterContext*, struct AVBufferSrcParameters*);

int dyn_av_buffersrc_parameters_set(struct AVFilterContext* p0, struct AVBufferSrcParameters* p1) {
    return _av_buffersrc_parameters_set(p0, p1);
};

static int (*_av_buffersrc_write_frame)(struct AVFilterContext*, struct AVFrame*);

int dyn_av_buffersrc_write_frame(struct AVFilterContext* p0, struct AVFrame* p1) {
    return _av_buffersrc_write_frame(p0, p1);
};

char *goav_load_avfilter() {
    char *ret;
    handle = dlopen("libavfilter.so", RTLD_NOW | RTLD_GLOBAL);
    if (ret = dlerror()) {
        return ret;
    }
    _avfilter_graph_config = dlsym(handle, "avfilter_graph_config");
    if (ret = dlerror()) {
        return ret;
    }
    _avfilter_graph_create_filter = dlsym(handle, "avfilter_graph_create_filter");
    if (ret = dlerror()) {
        return ret;
    }
    _avfilter_graph_free = dlsym(handle, "avfilter_graph_free");
    if (ret = dlerror()) {
        return ret;
    }
    _avfilter_inout_free = dlsym(handle, "avfilter_inout_free");
    if (ret = dlerror()) {
        return ret;
    }
    _av_buffersink_get_frame = dlsym(handle, "av_buffersink_get_frame");
    if (ret = dlerror()) {
        return ret;
    }
    _avfilter_get_by_name = dlsym(handle, "avfilter_get_by_name");
    if (ret = dlerror()) {
        return ret;
    }
    _avfilter_link = dlsym(handle, "avfilter_link");
    if (ret = dlerror()) {
        return ret;
    }
    _av_buffersrc_parameters_alloc = dlsym(handle, "av_buffersrc_parameters_alloc");
    if (ret = dlerror()) {
        return ret;
    }
    _avfilter_graph_alloc = dlsym(handle, "avfilter_graph_alloc");
    if (ret = dlerror()) {
        return ret;
    }
    _avfilter_graph_parse2 = dlsym(handle, "avfilter_graph_parse2");
    if (ret = dlerror()) {
        return ret;
    }
    _av_buffersrc_parameters_set = dlsym(handle, "av_buffersrc_parameters_set");
    if (ret = dlerror()) {
        return ret;
    }
    _av_buffersrc_write_frame = dlsym(handle, "av_buffersrc_write_frame");
    if (ret = dlerror()) {
        return ret;
    }
    return 0;
}
*/
import "C"

var (
	initOnce  sync.Once
	initError error
	initFuncs []func()
)

func dynamicInit() {
	initOnce.Do(func() {
		if ret := C.goav_load_avfilter(); ret != nil {
			initError = errors.Errorf("failed to initialize libavfilter: %s", C.GoString(ret))
		} else {
			for _, f := range initFuncs {
				f()
			}
		}
	})
	if initError != nil {
		panic(initError)
	}
}
func ConfigGraph(p0 *Graph, p1 unsafe.Pointer) int32 {
	dynamicInit()
	defer runtime.KeepAlive(p0)
	ret := C.dyn_avfilter_graph_config((*C.struct_AVFilterGraph)(unsafe.Pointer(p0)), p1)
	return *(*int32)(unsafe.Pointer(&ret))
}
func CreateFilterGraph(p0 **Context, p1 *Filter, p2 string, p3 string, p4 unsafe.Pointer, p5 *Graph) int32 {
	dynamicInit()
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
	ret := C.dyn_avfilter_graph_create_filter((**C.struct_AVFilterContext)(unsafe.Pointer(p0)), (*C.struct_AVFilter)(unsafe.Pointer(p1)), s2, s3, p4, (*C.struct_AVFilterGraph)(unsafe.Pointer(p5)))
	return *(*int32)(unsafe.Pointer(&ret))
}
func FreeGraph(p0 **Graph) {
	dynamicInit()
	defer runtime.KeepAlive(p0)
	C.dyn_avfilter_graph_free((**C.struct_AVFilterGraph)(unsafe.Pointer(p0)))
}
func FreeInOut(p0 **InOut) {
	dynamicInit()
	defer runtime.KeepAlive(p0)
	C.dyn_avfilter_inout_free((**C.struct_AVFilterInOut)(unsafe.Pointer(p0)))
}
func GetBufferSinkFrame(p0 *Context, p1 *avutil.Frame) int32 {
	dynamicInit()
	defer runtime.KeepAlive(p0)
	defer runtime.KeepAlive(p1)
	ret := C.dyn_av_buffersink_get_frame((*C.struct_AVFilterContext)(unsafe.Pointer(p0)), (*C.struct_AVFrame)(unsafe.Pointer(p1)))
	return *(*int32)(unsafe.Pointer(&ret))
}
func GetByName(p0 string) *Filter {
	dynamicInit()
	var s0 *C.char
	if p0 != "" {
		s0 = C.CString(p0)
		defer C.free(unsafe.Pointer(s0))
	}
	return (*Filter)(unsafe.Pointer(C.dyn_avfilter_get_by_name(s0)))
}
func Link(p0 *Context, p1 uint32, p2 *Context, p3 uint32) int32 {
	dynamicInit()
	defer runtime.KeepAlive(p0)
	defer runtime.KeepAlive(p1)
	defer runtime.KeepAlive(p2)
	defer runtime.KeepAlive(p3)
	ret := C.dyn_avfilter_link((*C.struct_AVFilterContext)(unsafe.Pointer(p0)), *(*C.uint)(unsafe.Pointer(&p1)), (*C.struct_AVFilterContext)(unsafe.Pointer(p2)), *(*C.uint)(unsafe.Pointer(&p3)))
	return *(*int32)(unsafe.Pointer(&ret))
}
func NewBufferSourceParameters() *BufferSourceParameters {
	dynamicInit()
	return (*BufferSourceParameters)(unsafe.Pointer(C.dyn_av_buffersrc_parameters_alloc()))
}
func NewGraph() *Graph {
	dynamicInit()
	return (*Graph)(unsafe.Pointer(C.dyn_avfilter_graph_alloc()))
}
func ParseGraph(p0 *Graph, p1 string, p2 **InOut, p3 **InOut) int32 {
	dynamicInit()
	defer runtime.KeepAlive(p0)
	var s1 *C.char
	if p1 != "" {
		s1 = C.CString(p1)
		defer C.free(unsafe.Pointer(s1))
	}
	defer runtime.KeepAlive(p2)
	defer runtime.KeepAlive(p3)
	ret := C.dyn_avfilter_graph_parse2((*C.struct_AVFilterGraph)(unsafe.Pointer(p0)), s1, (**C.struct_AVFilterInOut)(unsafe.Pointer(p2)), (**C.struct_AVFilterInOut)(unsafe.Pointer(p3)))
	return *(*int32)(unsafe.Pointer(&ret))
}
func SetBufferSourceParameters(p0 *Context, p1 *BufferSourceParameters) int32 {
	dynamicInit()
	defer runtime.KeepAlive(p0)
	defer runtime.KeepAlive(p1)
	ret := C.dyn_av_buffersrc_parameters_set((*C.struct_AVFilterContext)(unsafe.Pointer(p0)), (*C.struct_AVBufferSrcParameters)(unsafe.Pointer(p1)))
	return *(*int32)(unsafe.Pointer(&ret))
}
func WriteBufferSourceFrame(p0 *Context, p1 *avutil.Frame) int32 {
	dynamicInit()
	defer runtime.KeepAlive(p0)
	defer runtime.KeepAlive(p1)
	ret := C.dyn_av_buffersrc_write_frame((*C.struct_AVFilterContext)(unsafe.Pointer(p0)), (*C.struct_AVFrame)(unsafe.Pointer(p1)))
	return *(*int32)(unsafe.Pointer(&ret))
}
