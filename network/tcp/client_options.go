package tcp

import (
	"github.com/cute-angelia/go-game-utils/packet"
	"github.com/cute-angelia/go-game-utils/packet/ipacket"
	"time"
)

const (
	defaultClientDialAddr          = "127.0.0.1:3553"
	defaultClientDialTimeout       = time.Second * 5
	defaultClientHeartbeatInterval = time.Second * 10

	defaultClientPackerName = "due"
)

type ClientOption func(o *clientOptions)

type clientOptions struct {
	addr              string        // 地址
	timeout           time.Duration // 拨号超时时间，默认5s
	heartbeatInterval time.Duration // 心跳间隔时间，默认10s

	packer ipacket.Packer
}

func defaultClientOptions() *clientOptions {
	return &clientOptions{
		addr:              defaultClientDialAddr,
		timeout:           defaultClientDialTimeout,
		heartbeatInterval: defaultClientHeartbeatInterval,
		packer:            packet.GetDefaultPacker(defaultClientPackerName),
	}
}

// WithClientDialAddr 设置拨号地址
func WithClientDialAddr(addr string) ClientOption {
	return func(o *clientOptions) { o.addr = addr }
}

// WithClientDialTimeout 设置拨号超时时间
func WithClientDialTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) { o.timeout = timeout }
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
