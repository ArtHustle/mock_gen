// Package runtimefixture is the source interface for the shape-rich
// behavior suite (test/sample_behavior_test.go). It mirrors the
// demo runtime at ai/sample/runtime.go and adds Publish to cover
// unnamed multi-returns, so the resulting mock exercises every
// distinct codegen branch the template has.
//
// The generated mock lives in the sibling test/ package and is
// regenerable via `go generate ./test/dummy/runtime_fixture/`.
package runtimefixture

import (
	"context"
	"io"
	"time"
)

//go:generate go run ../../../cmd/mock_gen --async runtime.go

type Trigger struct {
	ID   string
	Name string
}

type Event struct {
	At      time.Time
	Topic   string
	Payload []byte
}

type Runtime interface {
	io.Closer

	Activate(ctx context.Context, t *Trigger) error
	Fetch(ctx context.Context, ids ...string) (events []Event, total int, err error)
	Publish(ctx context.Context, topic string, payload []byte) (Event, error)
	Stats() (active int, queued int, err error)
	History(window time.Duration) []Event
	Subscribe(ctx context.Context, topics []string, handler func(Event) error) (cancel func(), err error)
}
