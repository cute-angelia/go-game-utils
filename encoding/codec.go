package encoding

import (
	"github.com/cute-angelia/go-game-utils/encoding/json"
	"github.com/cute-angelia/go-game-utils/encoding/msgpack"
	"github.com/cute-angelia/go-game-utils/encoding/proto"
	"github.com/cute-angelia/go-game-utils/encoding/xml"
	"log"
)

var codecs = make(map[string]Codec)

func init() {
	Register(json.DefaultCodec)
	Register(proto.DefaultCodec)
	Register(xml.DefaultCodec)
	Register(msgpack.DefaultCodec)
}

type Codec interface {
	// Name 编解码器类型
	Name() string
	// Marshal 编码
	Marshal(v interface{}) ([]byte, error)
	// Unmarshal 解码
	Unmarshal(data []byte, v interface{}) error
}

// Register 注册编解码器
func Register(codec Codec) {
	if codec == nil {
		log.Fatal("can't register a invalid codec")
	}

	name := codec.Name()

	if name == "" {
		log.Fatal("can't register a codec without name")
	}

	if _, ok := codecs[name]; ok {
		log.Printf("the old %s codec will be overwritten", name)
	}

	codecs[name] = codec
}

// Invoke 调用编解码器
func Invoke(name string) Codec {
	codec, ok := codecs[name]
	if !ok {
		log.Fatalf("%s codec is not registered", name)
	}

	return codec
}
