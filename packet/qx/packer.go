package qx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"go-game-utils/packet"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"sync"
)

type Packer struct {
	opts             *options
	once             sync.Once
	readerSizePool   sync.Pool
	readerBufferPool sync.Pool
}

type NocopyReader interface {
	// Next returns a slice containing the next n bytes from the buffer,
	// advancing the buffer as if the bytes had been returned by Read.
	Next(n int) (p []byte, err error)

	// Peek returns the next n bytes without advancing the reader.
	Peek(n int) (buf []byte, err error)

	// Release the memory space occupied by all read slices.
	Release() (err error)

	Slice(n int) (r NocopyReader, err error)
}

func NewPacker(opts ...Option) *Packer {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	if o.bufferBytes < 0 {
		log.Fatalf("the number of buffer bytes must be greater than or equal to 0, and give %d", o.bufferBytes)
	}

	p := &Packer{opts: o}

	p.readerSizePool = sync.Pool{New: func() any { return make([]byte, defaultSizeBytes) }}
	p.readerBufferPool = sync.Pool{New: func() any {
		return make([]byte, defaultSizeBytes+o.bufferBytes)
	}}

	return p
}

// ReadMessage 读取消息
func (p *Packer) ReadMessage(reader interface{}) ([]byte, error) {
	switch r := reader.(type) {
	case NocopyReader:
		return p.nocopyReadMessage(r)
	case io.Reader:
		return p.copyReadMessage(r)
	default:
		return nil, errors.New("ErrInvalidReader")
	}
}

// 无拷贝读取消息
func (p *Packer) nocopyReadMessage(reader NocopyReader) ([]byte, error) {
	buf, err := reader.Peek(defaultSizeBytes)
	if err != nil {
		return nil, err
	}

	var size int32

	if p.opts.byteOrder == binary.BigEndian {
		size = int32(binary.BigEndian.Uint32(buf))
	} else {
		size = int32(binary.LittleEndian.Uint32(buf))
	}

	if size == 0 {
		return nil, nil
	}

	//n := int(defaultSizeBytes + size)
	n := int(size)

	r, err := reader.Slice(n)
	if err != nil {
		return nil, err
	}

	buf, err = r.Next(n)
	if err != nil {
		return nil, err
	}

	if err = reader.Release(); err != nil {
		return nil, err
	}

	return buf, nil
}

// 拷贝读取消息
func (p *Packer) copyReadMessage(reader io.Reader) ([]byte, error) {
	// 从 pool 获取一个 buffer
	buf := p.readerSizePool.Get().([]byte)
	// 确保归还前清空 buffer 并归还到 pool
	defer func() {
		clear(buf)
		p.readerSizePool.Put(buf)
	}()

	_, err := io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}

	var size int32

	if p.opts.byteOrder == binary.BigEndian {
		size = int32(binary.BigEndian.Uint32(buf))
	} else {
		size = int32(binary.LittleEndian.Uint32(buf))
	}

	if size == 0 {
		return nil, nil
	}

	// 组包
	// 从 pool 获取一个 buffer
	data := p.readerBufferPool.Get().([]byte)
	// 确保归还前清空 buffer 并归还到 pool
	defer func() {
		clear(data)
		p.readerBufferPool.Put(data)
	}()

	// 第一个 size
	smallSlice := data[:size]
	copy(smallSlice[:defaultSizeBytes], buf)

	_, err = io.ReadFull(reader, smallSlice[defaultSizeBytes:])
	if err != nil {
		return nil, err
	}

	return data, nil
}

// PackMessage 打包消息
func (p *Packer) PackMessage(messageIn packet.Message) ([]byte, error) {
	msg := messageIn.(*Message)

	if len(msg.data) > p.opts.bufferBytes {
		return nil, errors.New("ErrMessageTooLarge")
	}

	var (
		size = defaultSizeBytes + defaultMainIdBytes + defaultSubIdBytes + len(msg.data)
		buf  = &bytes.Buffer{}
	)

	if p.opts.isProto {
		buf.Grow(size + defaultClientAppendLength)
		head := Head{}
		head.Length = msg.length
		head.Mainid = msg.mainID
		head.Subid = msg.subID
		headData, _ := proto.Marshal(&head)
		buf.Write(headData)
		buf.Write(msg.data)
		return buf.Bytes(), nil
	} else {
		buf.Grow(size)
	}

	err := binary.Write(buf, p.opts.byteOrder, int32(size))
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, p.opts.byteOrder, msg.mainID)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, p.opts.byteOrder, msg.subID)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, p.opts.byteOrder, msg.data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnpackMessage 解包消息
func (p *Packer) UnpackMessage(data []byte) (packet.Message, error) {
	var (
		ln     = defaultSizeBytes + defaultMainIdBytes + defaultSubIdBytes
		reader = bytes.NewReader(data)
		size   uint32
	)

	msg := new(Message)

	if p.opts.isProto {
		ln += defaultClientAppendLength

		if len(data)-ln < 0 {
			return nil, errors.New("ErrInvalidMessage1")
		}

		head := Head{}
		err := proto.Unmarshal(data[:ln], &head)
		if err != nil {
			return msg, err
		}
		msg.length = head.GetLength()
		msg.mainID = head.GetMainid()
		msg.subID = head.GetSubid()
		msg.data = data[ln:]
		return msg, nil
	}

	if len(data)-ln < 0 {
		return nil, errors.New("ErrInvalidMessage1")
	}

	err := binary.Read(reader, p.opts.byteOrder, &size)
	if err != nil {
		return nil, err
	}

	if uint64(len(data)) != uint64(size) {
		return nil, errors.New("ErrInvalidMessage2")
	}

	msg.length = int32(size)

	var mainid int32
	if err = binary.Read(reader, p.opts.byteOrder, &mainid); err != nil {
		return nil, err
	} else {
		msg.mainID = mainid
	}

	var subid int32
	if err = binary.Read(reader, p.opts.byteOrder, &subid); err != nil {
		return nil, err
	} else {
		msg.subID = subid
	}

	msg.data = data[ln:]

	return msg, nil
}

// PackHeartbeat 打包心跳
// 这里不实现
func (p *Packer) PackHeartbeat() ([]byte, error) {
	return []byte{}, nil
}

// CheckHeartbeat 检测心跳包
// // 这里不实现
func (p *Packer) CheckHeartbeat(data []byte) (bool, error) {
	return true, nil
}
