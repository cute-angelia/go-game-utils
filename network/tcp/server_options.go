package tcp

import (
	"github.com/cute-angelia/go-game-utils/packet"
	"github.com/cute-angelia/go-game-utils/packet/ipacket"
	"time"
)

const (
	defaultServerAddr               = ":3553"
	defaultServerMaxConnNum         = 5000
	defaultServerHeartbeatInterval  = time.Second * 10
	defaultServerHeartbeatMechanism = "resp"

	defaultServerPackerName = "due"
)

const (
	RespHeartbeat HeartbeatMechanism = "resp" // 响应式心跳
	TickHeartbeat HeartbeatMechanism = "tick" // 主动定时心跳
)

type HeartbeatMechanism string

type ServerOption func(o *serverOptions)

type serverOptions struct {
	addr               string             // 监听地址，默认0.0.0.0:3553
	maxConnNum         int                // 最大连接数，默认5000
	heartbeatInterval  time.Duration      // 心跳检测间隔时间，默认10s
	heartbeatMechanism HeartbeatMechanism // 心跳机制，默认resp

	packer ipacket.Packer
}

func defaultServerOptions() *serverOptions {
	return &serverOptions{
		addr:               defaultServerAddr,
		maxConnNum:         defaultServerMaxConnNum,
		heartbeatInterval:  defaultServerHeartbeatInterval,
		heartbeatMechanism: HeartbeatMechanism(defaultServerHeartbeatMechanism),
		packer:             packet.GetDefaultPacker(defaultClientPackerName),
	}
}

// WithServerListenAddr 设置监听地址
func WithServerListenAddr(addr string) ServerOption {
	return func(o *serverOptions) { o.addr = addr }
}

// WithServerMaxConnNum 设置连接的最大连接数
func WithServerMaxConnNum(maxConnNum int) ServerOption {
	return func(o *serverOptions) { o.maxConnNum = maxConnNum }
}

// WithServerHeartbeatInterval 设置心跳检测间隔时间
func WithServerHeartbeatInterval(heartbeatInterval time.Duration) ServerOption {
	return func(o *serverOptions) { o.heartbeatInterval = heartbeatInterval }
}

// WithServerHeartbeatMechanism 设置心跳机制
func WithServerHeartbeatMechanism(heartbeatMechanism HeartbeatMechanism) ServerOption {
	return func(o *serverOptions) { o.heartbeatMechanism = heartbeatMechanism }
}

func WithServerPacker(packer ipacket.Packer) ServerOption {
	return func(o *serverOptions) {
		o.packer = packer
	}
}
