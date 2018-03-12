package channelio

import (
	"encoding/json"
	"io"
	"reflect"
)

// NewJSONEmitter instanciates a new Emitter that serializes in JSON.
func NewJSONEmitter(w io.Writer) Emitter {
	return jsonEmitter{
		Encoder: json.NewEncoder(w),
	}
}

type jsonEmitter struct {
	*json.Encoder
}

// Emit a value.
func (e jsonEmitter) Emit(value interface{}) error {
	return e.Encode(value)
}

// NewJSONReceiver instanciates a new Receiver that deserializes JSON.
//
// If a value is received that can't be properly deserialized as the specified
// value type, an error is returned.
func NewJSONReceiver(r io.Reader, valueType reflect.Type) Receiver {
	return jsonReceiver{
		Decoder:   json.NewDecoder(r),
		valueType: valueType,
		value:     reflect.New(valueType).Interface(),
	}
}

type jsonReceiver struct {
	*json.Decoder
	valueType reflect.Type
	value     interface{}
}

// Emit a value.
func (r jsonReceiver) Receive() (interface{}, error) {
	if err := r.Decode(r.value); err != nil {
		return nil, err
	}

	return reflect.ValueOf(r.value).Elem().Interface(), nil
}

// NewJSONTransmitter instanciates a new Transmitter that serializes and
// deserializes JSON.
//
// If a value is received that can't be properly deserialized as the specified
// value type, an error is returned.
func NewJSONTransmitter(r io.Reader, w io.Writer, valueType reflect.Type) Transmitter {
	return ComposeTransmitter(NewJSONEmitter(w), NewJSONReceiver(r, valueType))
}
