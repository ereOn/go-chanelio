package main

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"time"

	channelio "github.com/ereOn/go-channelio"
)

// A Person structure.
type Person struct {
	Name string
	Age  int
}

func main() {
	server, err := net.Listen("tcp", ":8000")

	if err != nil {
		panic(err)
	}

	defer server.Close()

	fmt.Printf("Waiting for an incoming connection...\n")

	serverConn, err := server.Accept()

	if err != nil {
		panic(err)
	}

	transmitter := channelio.NewJSONTransmitter(serverConn, serverConn, reflect.TypeOf(Person{}))

	delay := time.Second
	ctx, cancel := context.WithTimeout(context.Background(), delay)
	defer cancel()

	fmt.Printf("Waiting for %s\n", delay)

	emitterValues := make(chan interface{}, 1)
	receiverValues := make(chan interface{}, 1)

	go func() {
		for value := range receiverValues {
			emitterValues <- value
		}

		close(emitterValues)
	}()

	channelio.RunTransmitter(ctx, transmitter, emitterValues, receiverValues)

	close(receiverValues)
}
