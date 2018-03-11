package chanelio

import (
	"context"
	"io"
	"reflect"
	"testing"
)

func TestRunTransmitter(t *testing.T) {
	value := 42

	r, w := io.Pipe()
	transmitter := NewJSONTransmitter(r, w, reflect.TypeOf(value))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emitterValues := make(chan interface{}, 1)
	receiverValues := make(chan interface{}, 1)
	errors := make(chan error, 1)

	go func() {
		errors <- RunTransmitter(ctx, transmitter, emitterValues, receiverValues)
	}()

	emitterValues <- value
	x := <-receiverValues

	if y, ok := x.(int); !ok {
		t.Errorf("expected an %v, got a %v (%v)", reflect.TypeOf(y), reflect.TypeOf(x), x)
	} else if y != value {
		t.Errorf("expected %d got %d", value, y)
	}

	emitterValues <- value
	close(emitterValues)

	cancel()

	if err := <-errors; err == nil {
		t.Fatalf("expected an error")
	}
}

func BenchmarkRunTransmitter(b *testing.B) {
	value := 42

	r, w := io.Pipe()
	transmitter := NewJSONTransmitter(r, w, reflect.TypeOf(value))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emitterValues := make(chan interface{}, 1)
	receiverValues := make(chan interface{}, 1)
	errors := make(chan error, 1)

	go func() {
		errors <- RunTransmitter(ctx, transmitter, emitterValues, receiverValues)
	}()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		emitterValues <- value
		x := <-receiverValues

		if y, ok := x.(int); !ok {
			b.Errorf("expected an %v, got a %v (%v)", reflect.TypeOf(y), reflect.TypeOf(x), x)
		} else if y != value {
			b.Errorf("expected %d got %d", value, y)
		}
	}
}
