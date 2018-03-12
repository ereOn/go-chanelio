package channelio

import (
	"context"
	"io"
	"net"
	"reflect"
	"testing"
)

func TestRunTransmitter(t *testing.T) {
	value := 42

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, w := io.Pipe()
	transmitter := NewJSONTransmitter(r, w, reflect.TypeOf(value))

	go func() {
		<-ctx.Done()
		w.Close()
	}()

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

func TestRunTransmitterOverNetwork(t *testing.T) {
	value := 42

	server, err := net.Listen("tcp", ":0")

	if err != nil {
		t.Fatalf("failed to listen: %s", err)
	}

	defer server.Close()

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		serverConn, err := server.Accept()

		if err != nil {
			t.Fatalf("failed to accept: %s", err)
		}

		go func() {
			<-ctx.Done()
			serverConn.Close()
		}()

		transmitter := NewJSONTransmitter(serverConn, serverConn, reflect.TypeOf(value))

		emitterValues := make(chan interface{}, 1)
		receiverValues := make(chan interface{}, 1)

		go func() {
			for value := range receiverValues {
				emitterValues <- value
			}

			close(emitterValues)
		}()

		RunTransmitter(ctx, transmitter, emitterValues, receiverValues)

		close(receiverValues)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientConn, err := net.Dial("tcp", server.Addr().String())

	if err != nil {
		t.Fatalf("failed to connect: %s", err)
	}

	go func() {
		<-ctx.Done()
		clientConn.Close()
	}()

	defer clientConn.Close()

	transmitter := NewJSONTransmitter(clientConn, clientConn, reflect.TypeOf(value))

	emitterValues := make(chan interface{}, 1)
	receiverValues := make(chan interface{}, 1)

	go func() {
		RunTransmitter(ctx, transmitter, emitterValues, receiverValues)
	}()

	emitterValues <- value
	x := <-receiverValues

	if y, ok := x.(int); !ok {
		t.Errorf("expected an %v, got a %v (%v)", reflect.TypeOf(y), reflect.TypeOf(x), x)
	} else if y != value {
		t.Errorf("expected %d got %d", value, y)
	}
}

func BenchmarkRunTransmitter(b *testing.B) {
	value := 42

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, w := io.Pipe()
	transmitter := NewJSONTransmitter(r, w, reflect.TypeOf(value))

	go func() {
		<-ctx.Done()
		w.Close()
	}()

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
