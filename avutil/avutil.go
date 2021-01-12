//go:generate go run github.com/ssttevee/go-av/generate

package avutil

// #cgo pkg-config: libavutil
//
// #include <libavutil/avutil.h>
// #include <libavutil/buffer.h>
// #include <libavutil/dict.h>
// #include <libavutil/frame.h>
// #include <libavutil/pixdesc.h>
// #include <libavutil/hwcontext.h>
// #include <libavutil/log.h>
// #include <libavutil/opt.h>
import "C"

// +gen convtype struct_AVClass Class
// +gen convtype struct_AVFrame Frame
// +gen convtype struct_AVBuffer Buffer
// +gen convtype struct_AVBufferRef BufferRef
// +gen convtype struct_AVFrameSideData FrameSideData
// +gen convtype struct_AVDictionary Dictionary
// +gen convtype struct_AVRational Rational
// +gen convtype struct_AVOption Option
// +gen convtype struct_AVHWDeviceContext HWDeviceContext
// +gen convtype struct_AVHWFramesContext HWFramesContext

// +gen fieldtype struct_AVHWFramesContext free unsafe.Pointer
// +gen fieldtype struct_AVHWFramesContext format PixelFormat
// +gen fieldtype struct_AVHWFramesContext sw_format PixelFormat
// +gen fieldtype struct_AVOption _type OptionType

// +gen wrapfunc av_frame_alloc NewFrame
// +gen wrapfunc av_frame_free FreeFrame
// +gen wrapfunc av_frame_ref RefFrame
// +gen wrapfunc av_frame_unref UnrefFrame
// +gen wrapfunc av_frame_copy_props CopyFrameProps

// +gen wrapfunc av_buffer_ref RefBuffer
// +gen wrapfunc av_buffer_unref UnrefBuffer

// +gen wrapfunc av_dict_set SetDict
// +gen wrapfunc av_dict_free FreeDict

// +gen wrapfunc av_opt_set SetOpt
// +gen wrapfunc av_opt_set_int SetOptInt
// +gen wrapfunc av_opt_set_double SetOptDouble
// +gen wrapfunc av_opt_set_q SetOptRational
// +gen wrapfunc av_opt_set_pixel_fmt SetOptPixelFormat
// +gen wrapfunc av_opt_set_bin SetOptBin
// +gen wrapfunc av_opt_find2 FindOpt

// +gen wrapfunc av_hwdevice_ctx_create NewHWDeviceContext

// +gen wrapfunc av_hwframe_ctx_alloc NewHWFramesContext
// +gen wrapfunc av_hwframe_ctx_init InitHWFramesContext
// +gen wrapfunc av_hwframe_get_buffer GetHWFrameBuffer
// +gen wrapfunc av_hwframe_transfer_data TransferHWFrameData

// +gen wrapfunc av_rescale_rnd RescaleRound
// +gen wrapfunc av_mul_q MultiplyRational
// +gen wrapfunc av_strdup DupeString
// +gen wrapfunc av_free Free
// +gen wrapfunc av_malloc Malloc
// +gen wrapfunc av_get_pix_fmt_name getPixelFormatName
// +gen wrapfunc av_get_sample_fmt_name getSampleFormatName
// +gen wrapfunc av_hwdevice_get_type_name getHWDeviceTypeName
// +gen wrapfunc av_q2d q2d

// +gen paramtype av_hwdevice_ctx_create 1 HWDeviceType
// +gen paramtype av_opt_set_pixel_fmt 2 PixelFormat
// +gen paramtype av_get_pix_fmt_name 0 PixelFormat
// +gen paramtype av_get_sample_fmt_name 0 SampleFormat
// +gen paramtype av_hwdevice_get_type_name 0 HWDeviceType
