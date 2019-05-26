package pubee

import (
	"encoding/json"
	"errors"

	"github.com/gogo/protobuf/proto"
)

type MarshalFunc func(interface{}) ([]byte, error)

func MarshalDefault(in interface{}) ([]byte, error) {
	switch v := in.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return MarshalJSON(in)
	}
}

func MarshalJSON(in interface{}) ([]byte, error) {
	return json.Marshal(in)
}

func MarshalProtobuf(in interface{}) ([]byte, error) {
	if m, ok := in.(proto.Message); ok {
		return proto.Marshal(m)
	}
	return nil, errors.New("message should implement proto.Message interface")
}
