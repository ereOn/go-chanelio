package chanelio

import (
	"context"
	"io"
	"reflect"
	"testing"
)

func TestRun(t *testing.T) {
	r, w := io.Pipe()
	receiver := NewJSONReceiver(r, reflect.TypeOf(0))
	emitter := NewJSONEmitter(w)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emitterValues := make(chan interface{}, 1)
	receiverValues := make(chan interface{}, 1)
	errors := make(chan error, 1)

	go func() {
		errors <- Run(ctx, emitter, emitterValues, receiver, receiverValues)
	}()

	emitterValues <- 1
	x := <-receiverValues

	if y, ok := x.(int); !ok {
		t.Errorf("expected an %v, got a %v (%v)", reflect.TypeOf(y), reflect.TypeOf(x), x)
	} else if y != 1 {
		t.Errorf("expected %d got %d", 1, y)
	}

	emitterValues <- 2
	close(emitterValues)

	cancel()

	if err := <-errors; err == nil {
		t.Fatalf("expected an error")
	}
}
