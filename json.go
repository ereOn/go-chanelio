package chanelio

import (
	"context"
	"encoding/json"
	"io"
	"reflect"
)

// NewJSONEmitter implements a new JSONEmitter.
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

// NewJSONReceiver implements a new JSONReceiver.
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
