package qx

import (
	"encoding/binary"
	"go-game-utils/packet"
	"strings"
)

// Format: |--Length(4)--|--MainID(4)--|--SubID(4)--|--Data(variable)--|
const (
	defaultSizeBytes          = 4
	defaultMainIdBytes        = 4
	defaultSubIdBytes         = 4
	defaultClientAppendLength = 3

	defaultBufferBytes = 5000
	defaultEndian      = packet.LittleEndian
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

	// proto
	isProto bool
}

type Option func(o *options)

func defaultOptions() *options {
	opts := &options{
		byteOrder:   binary.BigEndian,
		bufferBytes: defaultBufferBytes,
	}

	endian := defaultEndian
	switch strings.ToLower(endian) {
	case packet.LittleEndian:
		opts.byteOrder = binary.LittleEndian
	case packet.BigEndian:
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

func WithIsProto(isProto bool) Option {
	return func(o *options) { o.isProto = isProto }
}

// WithEndian  大小端
func WithEndian(endian string) Option {
	return func(o *options) {
		o.endian = endian

		switch strings.ToLower(endian) {
		case packet.LittleEndian:
			o.byteOrder = binary.LittleEndian
		case packet.BigEndian:
			o.byteOrder = binary.BigEndian
		}
	}
}
