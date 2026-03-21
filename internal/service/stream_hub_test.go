package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSubscribe(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	assert.NotNil(t, ch)

	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestSubscribe_ContextCancel(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())

	ch := hub.Subscribe(ctx)
	cancel()
	time.Sleep(10 * time.Millisecond)

	_, ok := <-ch
	assert.False(t, ok, "channel should be closed after context cancel")
}

func TestPublish(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	hub.Publish(StreamEvent{
		Kind: "test",
		Msg:  "hello",
	})

	select {
	case ev := <-ch:
		assert.Equal(t, "test", ev.Kind)
		assert.Equal(t, "hello", ev.Msg)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestPublish_SingleSubscriber(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	hub.Publish(StreamEvent{Kind: "stdout", Msg: "line1"})
	hub.Publish(StreamEvent{Kind: "stdout", Msg: "line2"})

	ev1 := <-ch
	ev2 := <-ch

	assert.Equal(t, "line1", ev1.Msg)
	assert.Equal(t, "line2", ev2.Msg)
}

func TestPublish_MultipleSubscribers(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1 := hub.Subscribe(ctx)
	ch2 := hub.Subscribe(ctx)

	hub.Publish(StreamEvent{Kind: "test", Msg: "broadcast"})

	ev1 := <-ch1
	ev2 := <-ch2

	assert.Equal(t, "broadcast", ev1.Msg)
	assert.Equal(t, "broadcast", ev2.Msg)
}

func TestBackpressure_SlowClient(t *testing.T) {
	hub := NewStreamHub(2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	hub.Publish(StreamEvent{Msg: "msg1"})
	hub.Publish(StreamEvent{Msg: "msg2"})
	hub.Publish(StreamEvent{Msg: "msg3"})

	time.Sleep(50 * time.Millisecond)

	for i := 0; i < 2; i++ {
		<-ch
	}

	_, ok := <-ch
	assert.False(t, ok, "slow client should be disconnected")
}

func TestPublish_IDAndTS(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	hub.Publish(StreamEvent{Kind: "test", Msg: "msg"})

	select {
	case ev := <-ch:
		assert.True(t, ev.ID > 0, "ID should be auto-assigned")
		assert.True(t, ev.TS > 0, "TS should be auto-assigned")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestBackpressure_RemoveOnFull(t *testing.T) {
	hub := NewStreamHub(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	hub.Publish(StreamEvent{Msg: "msg1"})
	hub.Publish(StreamEvent{Msg: "msg2"})

	time.Sleep(50 * time.Millisecond)

	<-ch

	_, ok := <-ch
	assert.False(t, ok, "channel should be closed")
}
