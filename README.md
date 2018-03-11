[![Build Status](https://travis-ci.org/ereOn/go-channelio.svg?branch=master)](https://travis-ci.org/ereOn/go-channelio)
[![Coverage Status](https://coveralls.io/repos/github/ereOn/go-channelio/badge.svg?branch=master)](https://coveralls.io/github/ereOn/go-channelio?branch=master)

# ChannelIO

ChannelIO is a Go library that transforms channels in `io.Reader`, `io.Writer` and the other way around.

## Example

The typical case for ChannelIO is serialization of data over the network.

Here is a [client example](examples/client/client.go):

```go
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
	conn, err := net.Dial("tcp", "localhost:8000")

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Connection established.")

	transmitter := channelio.NewJSONTransmitter(conn, conn, reflect.TypeOf(Person{}))

	delay := time.Second
	ctx, cancel := context.WithTimeout(context.Background(), delay)
	defer cancel()

	fmt.Printf("Waiting for %s\n", delay)

	emitterValues := make(chan interface{}, 3)
	emitterValues <- Person{"alice", 20}
	emitterValues <- Person{"bob", 25}
	emitterValues <- Person{"chris", 30}

	receiverValues := make(chan interface{}, 3)

	channelio.RunTransmitter(ctx, transmitter, emitterValues, receiverValues)

	close(receiverValues)

	for person := range receiverValues {
		person := person.(Person)
		fmt.Println(person.Name, person.Age)
	}
}
```

And here is a [server example](examples/server/server.go) to play along:

```go
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
```
