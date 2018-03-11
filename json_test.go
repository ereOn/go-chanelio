package channelio

import (
	"context"
	"io"
	"reflect"
	"testing"
)

type sink struct {
	data []byte
}

func (s sink) Read(b []byte) (int, error) {
	return copy(b, s.data), io.EOF
}
func (sink) Write([]byte) (int, error) {
	return 0, nil
}
func (sink) Close() error { return nil }

func BenchmarkJSONEmitter(b *testing.B) {
	emitter := NewJSONEmitter(sink{})
	ctx := context.Background()
	value := 42

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if err := emitter.Emit(ctx, value); err != nil {
			b.Fatalf("expected no error but got: %s", err)
		}
	}
}

func BenchmarkJSONReceiver(b *testing.B) {
	receiver := NewJSONReceiver(sink{data: []byte("42")}, reflect.TypeOf(0))
	ctx := context.Background()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if value, err := receiver.Receive(ctx); err != nil {
			b.Fatalf("expected no error but got: %s", err)
		} else if value != 42 {
			b.Fatalf("expected %d error but got: %d", 42, value)
		}
	}
}
