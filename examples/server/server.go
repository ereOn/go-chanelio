package main

import (
	"context"
	"fmt"
	"net"
	"reflect"

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

	for {
		fmt.Printf("Waiting for incoming connections...\n")

		ctx, cancel := context.WithCancel(context.Background())

		serverConn, err := server.Accept()

		if err != nil {
			panic(err)
		}

		go func() {
			<-ctx.Done()
			serverConn.Close()
		}()

		transmitter := channelio.NewJSONTransmitter(serverConn, serverConn, reflect.TypeOf(Person{}))

		emitterValues := make(chan interface{}, 1)
		receiverValues := make(chan interface{}, 1)

		go func() {
			for value := range receiverValues {
				fmt.Println("Transmitting value:", value)
				emitterValues <- value
			}
		}()

		// As nothing cancels the context, the only possible way out of
		// RunTransmitter is an error on the transport. For instance, when the
		// client disconnects.
		channelio.RunTransmitter(ctx, transmitter, emitterValues, receiverValues)

		cancel()
	}
}
