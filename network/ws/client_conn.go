package ws

import (
	"errors"
	"github.com/cute-angelia/go-game-utils/network"
	"github.com/cute-angelia/go-game-utils/utils/icall"
	"github.com/cute-angelia/go-game-utils/utils/inet"
	"log"

	"github.com/gorilla/websocket"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type clientConn struct {
	rw                sync.RWMutex    // 锁
	id                int64           // 连接ID
	uid               int64           // 用户ID
	conn              *websocket.Conn // TCP源连接
	state             int32           // 连接状态
	client            *client         // 客户端
	chLowWrite        chan chWrite    // 低级队列
	chHighWrite       chan chWrite    // 优先队列
	lastHeartbeatTime int64           // 上次心跳时间
	done              chan struct{}   // 写入完成信号
	close             chan struct{}   // 关闭信号
}

var _ network.Conn = &clientConn{}

func newClientConn(id int64, conn *websocket.Conn, client *client) network.Conn {
	c := &clientConn{
		id:                id,
		conn:              conn,
		state:             int32(network.ConnOpened),
		client:            client,
		chLowWrite:        make(chan chWrite, 4096),
		chHighWrite:       make(chan chWrite, 1024),
		lastHeartbeatTime: time.Now().UnixNano(),
		done:              make(chan struct{}),
		close:             make(chan struct{}),
	}

	icall.Go(c.read)

	icall.Go(c.write)

	if c.client.connectHandler != nil {
		c.client.connectHandler(c)
	}

	return c
}

// ID 获取连接ID
func (c *clientConn) ID() int64 {
	return c.id
}

// UID 获取用户ID
func (c *clientConn) UID() int64 {
	return atomic.LoadInt64(&c.uid)
}

// Bind 绑定用户ID
func (c *clientConn) Bind(uid int64) {
	atomic.StoreInt64(&c.uid, uid)
}

// Unbind 解绑用户ID
func (c *clientConn) Unbind() {
	atomic.StoreInt64(&c.uid, 0)
}

// Send 发送消息（异步）
// 由于gorilla/websocket库不支持一个连接得并发读写，因而使用Send方法会导致使用写锁操作
// 建议使用Push方法替代Send
func (c *clientConn) Send(msg []byte) (err error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if err = c.checkState(); err != nil {
		return
	}

	c.chHighWrite <- chWrite{typ: dataPacket, msg: msg}

	return
}

// Push 发送消息（异步）
func (c *clientConn) Push(msg []byte) (err error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if err = c.checkState(); err != nil {
		return
	}

	c.chLowWrite <- chWrite{typ: dataPacket, msg: msg}

	return
}

// State 获取连接状态
func (c *clientConn) State() network.ConnState {
	return network.ConnState(atomic.LoadInt32(&c.state))
}

// Close 关闭连接（主动关闭）
func (c *clientConn) Close(force ...bool) error {
	if len(force) > 0 && force[0] {
		return c.forceClose()
	} else {
		return c.graceClose()
	}
}

// LocalIP 获取本地IP
func (c *clientConn) LocalIP() (string, error) {
	addr, err := c.LocalAddr()
	if err != nil {
		return "", err
	}

	return inet.ExtractIP(addr)
}

// LocalAddr 获取本地地址
func (c *clientConn) LocalAddr() (net.Addr, error) {
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
func (c *clientConn) RemoteIP() (string, error) {
	addr, err := c.RemoteAddr()
	if err != nil {
		return "", err
	}

	return inet.ExtractIP(addr)
}

// RemoteAddr 获取远端地址
func (c *clientConn) RemoteAddr() (net.Addr, error) {
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
func (c *clientConn) checkState() error {
	switch network.ConnState(atomic.LoadInt32(&c.state)) {
	case network.ConnHanged:
		return errors.New("ErrConnectionHanged")
	case network.ConnClosed:
		return errors.New("ErrConnectionClosed")
	default:
		return nil
	}
}

// 优雅关闭
func (c *clientConn) graceClose() error {
	if !atomic.CompareAndSwapInt32(&c.state, int32(network.ConnOpened), int32(network.ConnHanged)) {
		return errors.New("ErrConnectionNotOpened")
	}

	c.rw.RLock()
	c.chLowWrite <- chWrite{typ: closeSig}
	c.rw.RUnlock()

	<-c.done

	if !atomic.CompareAndSwapInt32(&c.state, int32(network.ConnHanged), int32(network.ConnClosed)) {
		return errors.New("ErrConnectionNotHanged")
	}

	c.rw.Lock()
	close(c.chLowWrite)
	close(c.chHighWrite)
	close(c.close)
	close(c.done)
	conn := c.conn
	c.conn = nil
	c.rw.Unlock()

	err := conn.Close()

	if c.client.disconnectHandler != nil {
		c.client.disconnectHandler(c)
	}

	return err
}

// 强制关闭
func (c *clientConn) forceClose() error {
	if !atomic.CompareAndSwapInt32(&c.state, int32(network.ConnOpened), int32(network.ConnClosed)) {
		if !atomic.CompareAndSwapInt32(&c.state, int32(network.ConnHanged), int32(network.ConnClosed)) {
			return errors.New("ErrConnectionClosed")
		}
	}

	c.rw.Lock()
	close(c.chLowWrite)
	close(c.chHighWrite)
	close(c.close)
	close(c.done)
	conn := c.conn
	c.conn = nil
	c.rw.Unlock()

	err := conn.Close()

	if c.client.disconnectHandler != nil {
		c.client.disconnectHandler(c)
	}

	return err
}

// 读取消息
func (c *clientConn) read() {
	conn := c.conn

	for {
		select {
		case <-c.close:
			return
		default:
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					if _, ok := err.(*websocket.CloseError); !ok {
						log.Printf("read message failed: %v", err)
					}
				}
				_ = c.forceClose()
				return
			}

			if msgType != websocket.BinaryMessage {
				continue
			}

			if c.client.opts.heartbeatInterval > 0 {
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

			// ignore empty packet
			if len(msg) == 0 {
				continue
			}

			// check heartbeat packet
			isHeartbeat, err := c.client.opts.packer.CheckHeartbeat(msg)
			if err != nil {
				log.Printf("[%s] check heartbeat message error: %v", c.client.opts.packer.String(), err)
				continue
			}

			// ignore heartbeat packet
			if isHeartbeat {
				continue
			}

			if c.client.receiveHandler != nil {
				c.client.receiveHandler(c, msg)
			}
		}
	}
}

// 写入消息
func (c *clientConn) write() {
	var (
		conn   = c.conn
		ticker *time.Ticker
	)

	log.Println(c.client.opts.heartbeatInterval, "💓时间", c.client.opts.heartbeatInterval > 0)

	if c.client.opts.heartbeatInterval > 0 {
		ticker = time.NewTicker(c.client.opts.heartbeatInterval)
		defer ticker.Stop()
	} else {
		ticker = &time.Ticker{C: make(chan time.Time, 1)}
	}

	for {
		select {
		case r, ok := <-c.chHighWrite:
			if !ok {
				return
			}

			if !c.doWrite(conn, r) {
				return
			}
		case <-ticker.C:
			if !c.doHandleHeartbeat(conn) {
				return
			}
		default:
			select {
			case r, ok := <-c.chHighWrite:
				if !ok {
					return
				}

				if !c.doWrite(conn, r) {
					return
				}
			case r, ok := <-c.chLowWrite:
				if !ok {
					return
				}

				if !c.doWrite(conn, r) {
					return
				}
			case <-ticker.C:
				if !c.doHandleHeartbeat(conn) {
					return
				}
			}
		}
	}
}

// 执行写入操作
func (c *clientConn) doWrite(conn *websocket.Conn, r chWrite) bool {
	if r.typ == closeSig {
		c.rw.RLock()
		c.done <- struct{}{}
		c.rw.RUnlock()
		return false
	}

	if c.isClosed() {
		return false
	}

	if r.typ == heartbeatPacket {
		if msg, err := c.client.opts.packer.PackHeartbeat(); err != nil {
			log.Printf("pack heartbeat message error: %v", err)
			return true
		} else {
			r.msg = msg
		}
	}

	if err := conn.WriteMessage(websocket.BinaryMessage, r.msg); err != nil {
		if !errors.Is(err, net.ErrClosed) {
			if _, ok := err.(*websocket.CloseError); !ok {
				log.Printf("write message error: %v", err)
			}
		}
	}

	return true
}

// 处理心跳
func (c *clientConn) doHandleHeartbeat(conn *websocket.Conn) bool {

	deadline := time.Now().Add(-2 * c.client.opts.heartbeatInterval).UnixNano()

	if atomic.LoadInt64(&c.lastHeartbeatTime) < deadline {
		log.Printf("connection heartbeat timeout, cid: %d", c.id)
		_ = c.forceClose()
		return false
	} else {
		if c.isClosed() {
			return false
		}

		if heartbeat, err := c.client.opts.packer.PackHeartbeat(); err != nil {
			log.Printf("pack heartbeat message error: %v", err)
		} else {

			if c.client.opts.packer.String() != "qx" {
				log.Println("发送心跳", heartbeat)
			}

			// send heartbeat packet
			if err := conn.WriteMessage(websocket.BinaryMessage, heartbeat); err != nil {
				log.Printf("write heartbeat message error: %v", err)
			}
		}
	}

	return true
}

// 是否已关闭
func (c *clientConn) isClosed() bool {
	return network.ConnState(atomic.LoadInt32(&c.state)) == network.ConnClosed
}
