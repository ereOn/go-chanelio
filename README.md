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
```

And here is a [server example](examples/server/server.go) to play along:

```go
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
```
