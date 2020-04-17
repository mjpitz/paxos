package encoding

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

var Proto = &Encoding{
	Marshal: func(entry interface{}) ([]byte, error) {
		val, ok := entry.(proto.Message)
		if !ok {
			return nil, fmt.Errorf("not of type proto.Message")
		}
		return proto.Marshal(val)
	},
	Unmarshal: func(bytes []byte, entry interface{}) error {
		val, ok := entry.(proto.Message)
		if !ok {
			return fmt.Errorf("not of type proto.Message")
		}
		return proto.Unmarshal(bytes, val)
	},
}
