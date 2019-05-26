package marshal

import (
	"encoding/json"
	"errors"

	"github.com/golang/protobuf/proto"
)

type Func func(interface{}) ([]byte, error)

func Default(in interface{}) ([]byte, error) {
	switch v := in.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return JSON(in)
	}
}

func JSON(in interface{}) ([]byte, error) {
	return json.Marshal(in)
}

func Protobuf(in interface{}) ([]byte, error) {
	if m, ok := in.(proto.Message); ok {
		return proto.Marshal(m)
	}
	return nil, errors.New("message should implement proto.Message interface")
}
