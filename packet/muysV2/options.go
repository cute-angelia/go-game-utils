package muysV2

import (
	"encoding/binary"
	"github.com/cute-angelia/go-game-utils/encoding"
	"strings"
)

const Name = "muysV2"

const (
	LittleEndian = "little"
	BigEndian    = "big"
)

// Format: |--Length(4)--|--MsgType(2)--|--Data(variable)--|
const (
	defaultSizeBytes = 4 // uint32
	defaultTypeBytes = 2 // uint16

	defaultBufferBytes = 5000
	defaultEndian      = BigEndian
)

type options struct {
	// 字节序
	// 默认为binary.LittleEndian
	byteOrder binary.ByteOrder

	// 消息字节数
	// 默认为5000字节
	bufferBytes int

	// 大小端
	endian string

	// 编码器
	codeC encoding.Codec
}

type Option func(o *options)

func defaultOptions() *options {
	opts := &options{
		byteOrder:   binary.BigEndian,
		bufferBytes: defaultBufferBytes,
		codeC:       nil,
	}

	endian := defaultEndian
	switch strings.ToLower(endian) {
	case LittleEndian:
		opts.byteOrder = binary.LittleEndian
	case BigEndian:
		opts.byteOrder = binary.BigEndian
	}

	return opts
}

// WithByteOrder 设置字节序
func WithByteOrder(byteOrder binary.ByteOrder) Option {
	return func(o *options) { o.byteOrder = byteOrder }
}

// WithBufferBytes 设置消息字节数
func WithBufferBytes(bufferBytes int) Option {
	return func(o *options) { o.bufferBytes = bufferBytes }
}

// WithEndian  大小端
func WithEndian(endian string) Option {
	return func(o *options) {
		o.endian = endian

		switch strings.ToLower(endian) {
		case LittleEndian:
			o.byteOrder = binary.LittleEndian
		case BigEndian:
			o.byteOrder = binary.BigEndian
		}
	}
}

func WithCodeC(codecName string) Option {
	return func(o *options) {
		switch codecName {
		case "proto":
			o.codeC = encoding.Invoke("proto")
		case "json":
			o.codeC = encoding.Invoke("json")
		case "msgpack":
			o.codeC = encoding.Invoke("msgpack")
		case "xml":
			o.codeC = encoding.Invoke("xml")
		default:
			o.codeC = nil
		}
	}
}
