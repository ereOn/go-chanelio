package channelio

import (
	"context"
)

// Emitter represents a type that is able to encode a given value.
type Emitter interface {
	// Emit a value.
	//
	// If the specified context expires, the returned error must be the
	// context's error.
	Emit(ctx context.Context, value interface{}) error
}

// Receiver represents a type that is able to decode a given value.
type Receiver interface {
	// Receive a value.
	//
	// If the specified context expires, the returned error must be the
	// context's error.
	Receive(ctx context.Context) (interface{}, error)
}

// Transmitter represents a type that acts both as a Receiver and an Emitter.
type Transmitter interface {
	Emitter
	Receiver
}

// RunEmitter reads all the values from the specified channel and pushes
// them through the specified Emitter.
//
// The call only returns if either:
// - The specified context expires. In that case the context error is returned.
// - The emitting process returns an error. In that case, this error is
// returned.
//
// If the values channel is closed, the call will still block until the
// specified context expires. To control the lifetime of the call, the caller
// must control the expiration of the context.
//
// The caller may close the channel to indicate that no more values are to be
// emitted. Note that even in that case, the call will still block until the
// specified context expires.
func RunEmitter(ctx context.Context, emitter Emitter, values <-chan interface{}) error {
	for {
		// This is necessary over a range statement: if the values channel is
		// empty, we must still honor the context possibly expiring.
		select {
		case value, ok := <-values:
			if !ok {
				values = nil
				break
			}

			if err := emitter.Emit(ctx, value); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// RunReceiver continuously reads values from the specified Receiver and pushes
// them to the specified channel.
//
// The call only returns if either:
// - The specified context expires. In that case the context error is returned.
// - The receiving process returns an error. In that case, this error is
// returned.
//
// The caller must not close the channel while the call is executing.
func RunReceiver(ctx context.Context, receiver Receiver, values chan<- interface{}) error {
	for {
		value, err := receiver.Receive(ctx)

		if err != nil {
			return err
		}

		// This is necessary: if the values channel is not able to receive
		// the value, we must still honor the context possibly expiring.
		select {
		case values <- value:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// RunTransmitter combines the RunEmitter and RunReceiver functions.
//
// The call only returns if either:
// - The specified context expires. In that case the context error is returned.
// - The emitting process returns an error. In that case, this error is
// returned.
// - The receiving process returns an error. In that case, this error is
// returned.
//
// The caller may close the emitter channel to indicate that no more values are
// to be emitted. Note that even in that case, the call will still block until
// the specified context expires or the receiving process fails.
//
// The caller must not close the receiver channel while the call is executing.
func RunTransmitter(ctx context.Context, transmitter Transmitter, emitterValues <-chan interface{}, receiverValues chan<- interface{}) error {
	ctx, cancel := context.WithCancel(ctx)

	// We make sure both our coroutines don't stay blocked forever on trying to
	// write their result.
	//
	// Whichever results comes through first wins.
	result := make(chan error, 2)

	go func() {
		result <- RunEmitter(ctx, transmitter, emitterValues)
	}()

	go func() {
		result <- RunReceiver(ctx, transmitter, receiverValues)
	}()

	// We get the first result which will be the return value of the call.
	err := <-result

	// Force the other goroutine to unblock and wait for it to finish. This is
	// to ensure we don't leave the call with a pending RunEmitter call that
	// could panic if its channel gets closed.
	cancel()
	<-result

	return err
}

// ComposeTransmitter composes an Emitter and a Receiver into a Transmitter.
func ComposeTransmitter(emitter Emitter, receiver Receiver) Transmitter {
	return transmitter{
		Emitter:  emitter,
		Receiver: receiver,
	}
}

type transmitter struct {
	Emitter
	Receiver
}
