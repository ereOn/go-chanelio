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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := net.Dial("tcp", "localhost:8000")

	if err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	fmt.Println("Connection established.")

	transmitter := channelio.NewJSONTransmitter(conn, conn, reflect.TypeOf(Person{}))

	emitterValues := make(chan interface{}, 3)
	emitterValues <- Person{"alice", 20}
	emitterValues <- Person{"bob", 25}
	emitterValues <- Person{"chris", 30}

	receiverValues := make(chan interface{}, 3)

	go func() {
		fmt.Println("Received:", (<-receiverValues).(Person))
		fmt.Println("Received:", (<-receiverValues).(Person))
		fmt.Println("Received:", (<-receiverValues).(Person))

		// We received everything we expected: time to cancel the context and
		// cause the blocking call to RunTransmitter below to complete.
		cancel()
	}()

	channelio.RunTransmitter(ctx, transmitter, emitterValues, receiverValues)
}
