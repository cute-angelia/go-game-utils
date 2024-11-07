package tcp

import (
	"errors"
	"github.com/cute-angelia/go-game-utils/network"
	"github.com/cute-angelia/go-game-utils/utils/icall"
	"github.com/cute-angelia/go-game-utils/utils/inet"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type serverConn struct {
	id                int64          // 连接ID
	uid               int64          // 用户ID
	state             int32          // 连接状态
	connMgr           *serverConnMgr // 连接管理
	rw                sync.RWMutex   // 读写锁
	conn              net.Conn       // TCP源连接
	chWrite           chan chWrite   // 写入队列
	done              chan struct{}  // 写入完成信号
	close             chan struct{}  // 关闭信号
	lastHeartbeatTime int64          // 上次心跳时间
}

var _ network.Conn = &serverConn{}

// ID 获取连接ID
func (c *serverConn) ID() int64 {
	return c.id
}

// UID 获取用户ID
func (c *serverConn) UID() int64 {
	return atomic.LoadInt64(&c.uid)
}

// Bind 绑定用户ID
func (c *serverConn) Bind(uid int64) {
	atomic.StoreInt64(&c.uid, uid)
}

// Unbind 解绑用户ID
func (c *serverConn) Unbind() {
	atomic.StoreInt64(&c.uid, 0)
}

// Send 发送消息（同步）
func (c *serverConn) Send(msg []byte) (err error) {
	if err = c.checkState(); err != nil {
		return
	}

	c.rw.RLock()
	conn := c.conn
	c.rw.RUnlock()

	if conn == nil {
		return errors.New("ErrConnectionClosed")
	}

	_, err = conn.Write(msg)
	return
}

// Push 发送消息（异步）
func (c *serverConn) Push(msg []byte) (err error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if err = c.checkState(); err != nil {
		return
	}

	c.chWrite <- chWrite{typ: dataPacket, msg: msg}

	return
}

// State 获取连接状态
func (c *serverConn) State() network.ConnState {
	return network.ConnState(atomic.LoadInt32(&c.state))
}

// Close 关闭连接
func (c *serverConn) Close(force ...bool) error {
	if len(force) > 0 && force[0] {
		return c.forceClose(true)
	} else {
		return c.graceClose(true)
	}
}

// LocalIP 获取本地IP
func (c *serverConn) LocalIP() (string, error) {
	addr, err := c.LocalAddr()
	if err != nil {
		return "", err
	}

	return inet.ExtractIP(addr)
}

// LocalAddr 获取本地地址
func (c *serverConn) LocalAddr() (net.Addr, error) {
	if err := c.checkState(); err != nil {
		return nil, err
	}

	c.rw.RLock()
	conn := c.conn
	c.rw.RUnlock()

	if conn == nil {
		return nil, errors.New("ErrConnectionClosed")
	}

	return conn.LocalAddr(), nil
}

// RemoteIP 获取远端IP
func (c *serverConn) RemoteIP() (string, error) {
	addr, err := c.RemoteAddr()
	if err != nil {
		return "", err
	}

	return inet.ExtractIP(addr)
}

// RemoteAddr 获取远端地址
func (c *serverConn) RemoteAddr() (net.Addr, error) {
	if err := c.checkState(); err != nil {
		return nil, err
	}

	c.rw.RLock()
	conn := c.conn
	c.rw.RUnlock()

	if conn == nil {
		return nil, errors.New("ErrConnectionClosed")
	}

	return conn.RemoteAddr(), nil
}

// 检测连接状态
func (c *serverConn) checkState() error {
	switch network.ConnState(atomic.LoadInt32(&c.state)) {
	case network.ConnHanged:
		return errors.New("ErrConnectionHanged")
	case network.ConnClosed:
		return errors.New("ErrConnectionClosed")
	default:
		return nil
	}
}

// 初始化连接
func (c *serverConn) init(cm *serverConnMgr, id int64, conn net.Conn) {
	c.id = id
	c.conn = conn
	c.connMgr = cm
	c.chWrite = make(chan chWrite, 4096)
	c.done = make(chan struct{})
	c.close = make(chan struct{})
	c.lastHeartbeatTime = time.Now().UnixNano()
	atomic.StoreInt64(&c.uid, 0)
	atomic.StoreInt32(&c.state, int32(network.ConnOpened))

	icall.Go(c.read)

	icall.Go(c.write)

	if c.connMgr.server.connectHandler != nil {
		c.connMgr.server.connectHandler(c)
	}
}

// 优雅关闭
func (c *serverConn) graceClose(isNeedRecycle bool) error {
	if !atomic.CompareAndSwapInt32(&c.state, int32(network.ConnOpened), int32(network.ConnHanged)) {
		return errors.New("ErrConnectionNotOpened")
	}

	c.rw.RLock()
	c.chWrite <- chWrite{typ: closeSig}
	c.rw.RUnlock()

	<-c.done

	if !atomic.CompareAndSwapInt32(&c.state, int32(network.ConnHanged), int32(network.ConnClosed)) {
		return errors.New("ErrConnectionNotHanged")
	}

	c.rw.Lock()
	close(c.chWrite)
	close(c.close)
	close(c.done)
	conn := c.conn
	c.conn = nil
	c.rw.Unlock()

	err := conn.Close()

	if isNeedRecycle {
		c.connMgr.recycle(conn)
	}

	if c.connMgr.server.disconnectHandler != nil {
		c.connMgr.server.disconnectHandler(c)
	}

	return err
}

// 强制关闭
func (c *serverConn) forceClose(isNeedRecycle bool) error {
	if !atomic.CompareAndSwapInt32(&c.state, int32(network.ConnOpened), int32(network.ConnClosed)) {
		if !atomic.CompareAndSwapInt32(&c.state, int32(network.ConnHanged), int32(network.ConnClosed)) {
			return errors.New("ErrConnectionClosed")
		}
	}

	c.rw.Lock()
	close(c.chWrite)
	close(c.close)
	close(c.done)
	conn := c.conn
	c.conn = nil
	c.rw.Unlock()

	err := conn.Close()

	if isNeedRecycle {
		c.connMgr.recycle(conn)
	}

	if c.connMgr.server.disconnectHandler != nil {
		c.connMgr.server.disconnectHandler(c)
	}

	return err
}

// 读取消息
func (c *serverConn) read() {
	conn := c.conn

	for {
		select {
		case <-c.close:
			return
		default:
			msg, err := c.connMgr.server.opts.packer.ReadMessage(conn)
			if err != nil {
				_ = c.forceClose(true)
				return
			}

			if c.connMgr.server.opts.heartbeatInterval > 0 {
				atomic.StoreInt64(&c.lastHeartbeatTime, time.Now().UnixNano())
			}

			switch c.State() {
			case network.ConnHanged:
				continue
			case network.ConnClosed:
				return
			default:
				// ignore
			}

			isHeartbeat, err := c.connMgr.server.opts.packer.CheckHeartbeat(msg)
			if err != nil {
				log.Printf("check heartbeat message error: %v", err)
				continue
			}

			// ignore heartbeat packet
			if isHeartbeat {
				// responsive heartbeat
				if c.connMgr.server.opts.heartbeatMechanism == RespHeartbeat {
					if heartbeat, err := c.connMgr.server.opts.packer.PackHeartbeat(); err != nil {
						log.Printf("pack heartbeat message error: %v", err)
					} else {
						if _, err = conn.Write(heartbeat); err != nil {
							log.Printf("write heartbeat message error: %v", err)
						}
					}
				}
				continue
			}

			// ignore empty packet
			if len(msg) == 0 {
				continue
			}

			if c.connMgr.server.receiveHandler != nil {
				c.connMgr.server.receiveHandler(c, msg)
			}
		}
	}
}

// 写入消息
func (c *serverConn) write() {
	var (
		conn   = c.conn
		ticker *time.Ticker
	)

	if c.connMgr.server.opts.heartbeatInterval > 0 {
		ticker = time.NewTicker(c.connMgr.server.opts.heartbeatInterval)
		defer ticker.Stop()
	} else {
		ticker = &time.Ticker{C: make(chan time.Time, 1)}
	}

	for {
		select {
		case r, ok := <-c.chWrite:
			if !ok {
				return
			}

			if r.typ == closeSig {
				c.rw.RLock()
				c.done <- struct{}{}
				c.rw.RUnlock()
				return
			}

			if c.isClosed() {
				return
			}

			if _, err := conn.Write(r.msg); err != nil {
				log.Printf("write data message error: %v", err)
			}
		case <-ticker.C:
			deadline := time.Now().Add(-2 * c.connMgr.server.opts.heartbeatInterval).UnixNano()
			if atomic.LoadInt64(&c.lastHeartbeatTime) < deadline {
				log.Printf("connection heartbeat timeout, cid: %d", c.id)
				_ = c.forceClose(true)
				return
			} else {
				if c.connMgr.server.opts.heartbeatMechanism == TickHeartbeat {
					if c.isClosed() {
						return
					}

					if heartbeat, err := c.connMgr.server.opts.packer.PackHeartbeat(); err != nil {
						log.Printf("pack heartbeat message error: %v", err)
					} else {
						// send heartbeat packet
						if _, err = conn.Write(heartbeat); err != nil {
							log.Printf("write heartbeat message error: %v", err)
						}
					}
				}
			}
		}
	}
}

// 是否已关闭
func (c *serverConn) isClosed() bool {
	return network.ConnState(atomic.LoadInt32(&c.state)) == network.ConnClosed
}
