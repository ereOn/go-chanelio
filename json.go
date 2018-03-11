package channelio

import (
	"context"
	"encoding/json"
	"io"
	"reflect"
)

// NewJSONEmitter instanciates a new Emitter that serializes in JSON.
func NewJSONEmitter(w io.WriteCloser) Emitter {
	return jsonEmitter{
		Encoder: json.NewEncoder(w),
		Closer:  w,
	}
}

type jsonEmitter struct {
	*json.Encoder
	io.Closer
}

// Emit a value.
func (e jsonEmitter) Emit(ctx context.Context, value interface{}) error {
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			e.Close()
		case <-done:
		}
	}()

	return e.Encode(value)
}

// NewJSONReceiver instanciates a new Receiver that deserializes JSON.
//
// If a value is received that can't be properly deserialized as the specified
// value type, an error is returned.
func NewJSONReceiver(r io.ReadCloser, valueType reflect.Type) Receiver {
	return jsonReceiver{
		Decoder:   json.NewDecoder(r),
		Closer:    r,
		valueType: valueType,
		value:     reflect.New(valueType).Interface(),
	}
}

type jsonReceiver struct {
	*json.Decoder
	io.Closer
	valueType reflect.Type
	value     interface{}
}

// Emit a value.
func (r jsonReceiver) Receive(ctx context.Context) (interface{}, error) {
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			r.Close()
		case <-done:
		}
	}()

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
func NewJSONTransmitter(r io.ReadCloser, w io.WriteCloser, valueType reflect.Type) Transmitter {
	return ComposeTransmitter(NewJSONEmitter(w), NewJSONReceiver(r, valueType))
}
