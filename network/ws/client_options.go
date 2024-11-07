package ws

import (
	"github.com/cute-angelia/go-game-utils/packet"
	"github.com/cute-angelia/go-game-utils/packet/ipacket"
	"time"
)

const (
	defaultClientDialUrl           = "ws://127.0.0.1:3553"
	defaultClientHandshakeTimeout  = time.Second * 10
	defaultClientHeartbeatInterval = time.Second * 10

	defaultClientPackerName = "due"
)

type ClientOption func(o *clientOptions)

type clientOptions struct {
	url               string        // 拨号地址
	msgType           string        // 默认消息类型，text | binary
	handshakeTimeout  time.Duration // 握手超时时间
	heartbeatInterval time.Duration // 心跳间隔时间，默认10s

	packer ipacket.Packer
}

func defaultClientOptions() *clientOptions {
	return &clientOptions{
		url:               defaultClientDialUrl,
		handshakeTimeout:  defaultClientHandshakeTimeout,
		heartbeatInterval: defaultClientHeartbeatInterval,

		packer: packet.GetDefaultPacker(defaultClientPackerName),
	}
}

// WithClientDialUrl 设置拨号链接
func WithClientDialUrl(url string) ClientOption {
	return func(o *clientOptions) { o.url = url }
}

// WithClientHandshakeTimeout 设置握手超时时间
func WithClientHandshakeTimeout(handshakeTimeout time.Duration) ClientOption {
	return func(o *clientOptions) { o.handshakeTimeout = handshakeTimeout }
}

// WithClientHeartbeatInterval 设置心跳间隔时间
func WithClientHeartbeatInterval(heartbeatInterval time.Duration) ClientOption {
	return func(o *clientOptions) { o.heartbeatInterval = heartbeatInterval }
}

func WithClientPacker(packer ipacket.Packer) ClientOption {
	return func(o *clientOptions) {
		o.packer = packer
	}
}
