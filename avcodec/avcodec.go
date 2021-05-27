//go:generate go run github.com/ssttevee/go-av/generate

package avcodec

// #include <libavcodec/avcodec.h>
// #include <libavcodec/packet.h>
import "C"

// +gen convtype struct_AVDictionary github.com/ssttevee/go-av/avutil.Dictionary
// +gen convtype struct_AVFrame github.com/ssttevee/go-av/avutil.Frame
// +gen convtype struct_AVBufferRef github.com/ssttevee/go-av/avutil.BufferRef
// +gen convtype struct_AVRational github.com/ssttevee/go-av/avutil.Rational
// +gen convtype struct_AVClass github.com/ssttevee/go-av/avutil.Class

// +gen convtype struct_AVPacket Packet
// +gen convtype struct_AVCodecContext Context
// +gen convtype struct_AVCodec Codec
// +gen convtype struct_AVCodecParameters Parameters
// +gen convtype struct_AVBitStreamFilter BitstreamFilter
// +gen convtype struct_AVBSFContext BitstreamFilterContext

// +gen fieldtype struct_AVCodec id ID
// +gen fieldtype struct_AVCodec pix_fmts *github.com/ssttevee/go-av/avutil.PixelFormat
// +gen fieldtype struct_AVCodec sample_fmts *github.com/ssttevee/go-av/avutil.SampleFormat
// +gen fieldtype struct_AVCodec _type github.com/ssttevee/go-av/avutil.MediaType

// +gen fieldtype struct_AVCodecContext codec_type github.com/ssttevee/go-av/avutil.MediaType
// +gen fieldtype struct_AVCodecContext pix_fmt github.com/ssttevee/go-av/avutil.PixelFormat
// +gen fieldtype struct_AVCodecContext sample_fmt github.com/ssttevee/go-av/avutil.SampleFormat

// +gen fieldtype struct_AVCodecParameters codec_id ID

// +gen wrapfunc avcodec_open2 Open
// +gen wrapfunc avcodec_alloc_context3 NewContext
// +gen wrapfunc avcodec_free_context FreeContext
// +gen wrapfunc avcodec_parameters_to_context ParametersToContext
// +gen wrapfunc avcodec_parameters_from_context ParametersFromContext
// +gen wrapfunc avcodec_find_decoder_by_name FindDecoderByName
// +gen wrapfunc avcodec_find_encoder_by_name FindEncoderByName
// +gen wrapfunc avcodec_send_packet SendPacket
// +gen wrapfunc avcodec_receive_packet ReceivePacket
// +gen wrapfunc avcodec_send_frame SendFrame
// +gen wrapfunc avcodec_receive_frame ReceiveFrame
// +gen wrapfunc avcodec_parameters_copy CopyParameters
// +gen wrapfunc avcodec_get_name getName
// +gen wrapfunc avcodec_find_decoder FindDecoder
// +gen wrapfunc avcodec_find_encoder FindEncoder

// +gen wrapfunc av_packet_alloc NewPacket
// +gen wrapfunc av_packet_free FreePacket
// +gen wrapfunc av_packet_ref RefPacket
// +gen wrapfunc av_packet_unref UnrefPacket

// +gen wrapfunc av_bsf_alloc NewBitstreamFilter
// +gen wrapfunc av_bsf_free FreeBitstreamFilter
// +gen wrapfunc av_bsf_init InitBitstreamFilter
// +gen wrapfunc av_bsf_receive_packet ReceiveBitstreamFilterPacket
// +gen wrapfunc av_bsf_send_packet SendBitstreamFilterPacket
// +gen wrapfunc av_bsf_get_by_name GetBitstreamFilterByName

// +gen paramtype avcodec_get_name 0 ID
// +gen paramtype avcodec_find_decoder 0 ID
// +gen paramtype avcodec_find_encoder 0 ID
