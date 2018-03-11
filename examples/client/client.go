package main

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"time"

	chanelio "github.com/ereOn/go-chanelio"
)

// A Person structure.
type Person struct {
	Name string
	Age  int
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Connection established.")

	transmitter := chanelio.NewJSONTransmitter(conn, conn, reflect.TypeOf(Person{}))

	delay := time.Second
	ctx, cancel := context.WithTimeout(context.Background(), delay)
	defer cancel()

	fmt.Printf("Waiting for %s\n", delay)

	emitterValues := make(chan interface{}, 3)
	emitterValues <- Person{"alice", 20}
	emitterValues <- Person{"bob", 25}
	emitterValues <- Person{"chris", 30}

	receiverValues := make(chan interface{}, 3)

	chanelio.RunTransmitter(ctx, transmitter, emitterValues, receiverValues)

	close(receiverValues)

	for person := range receiverValues {
		person := person.(Person)
		fmt.Println(person.Name, person.Age)
	}
}
