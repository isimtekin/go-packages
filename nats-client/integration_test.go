//go:build integration

package natsclient

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

// Integration tests require a running NATS server with JetStream enabled.
// Run with: make test-integration
// Or manually: docker compose up -d && go test -v -tags=integration ./...

func TestIntegration_Connection(t *testing.T) {
	client, err := NewWithOptions(
		WithURL("nats://localhost:4222"),
		WithName("integration-test"),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	if !client.IsConnected() {
		t.Error("Client should be connected")
	}

	stats := client.Stats()
	t.Logf("Connection stats: InMsgs=%d, OutMsgs=%d", stats.InMsgs, stats.OutMsgs)
}

func TestIntegration_PubSub(t *testing.T) {
	client, err := NewWithOptions(
		WithURL("nats://localhost:4222"),
		WithName("pubsub-test"),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	received := make(chan *nats.Msg, 1)
	subject := "test.pubsub"

	// Subscribe
	sub, err := client.Subscribe(subject, func(msg *nats.Msg) {
		received <- msg
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer sub.Unsubscribe()

	// Publish
	testData := []byte("hello world")
	if err := client.Publish(subject, testData); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait for message
	select {
	case msg := <-received:
		if string(msg.Data) != string(testData) {
			t.Errorf("Data mismatch: got %s, want %s", msg.Data, testData)
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestIntegration_RequestReply(t *testing.T) {
	client, err := NewWithOptions(
		WithURL("nats://localhost:4222"),
		WithName("reqreply-test"),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	subject := "test.request"
	response := []byte("pong")

	// Setup responder
	sub, err := client.Subscribe(subject, func(msg *nats.Msg) {
		msg.Respond(response)
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer sub.Unsubscribe()

	// Send request
	reply, err := client.Request(subject, []byte("ping"), 5*time.Second)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if string(reply.Data) != string(response) {
		t.Errorf("Response mismatch: got %s, want %s", reply.Data, response)
	}
}

func TestIntegration_QueueSubscribe(t *testing.T) {
	client, err := NewWithOptions(
		WithURL("nats://localhost:4222"),
		WithName("queue-test"),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	subject := "test.queue"
	queue := "workers"
	received := make(chan *nats.Msg, 10)

	// Create multiple queue subscribers
	for i := 0; i < 3; i++ {
		sub, err := client.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
			received <- msg
		})
		if err != nil {
			t.Fatalf("QueueSubscribe failed: %v", err)
		}
		defer sub.Unsubscribe()
	}

	// Publish messages
	for i := 0; i < 5; i++ {
		if err := client.Publish(subject, []byte("msg")); err != nil {
			t.Fatalf("Publish failed: %v", err)
		}
	}

	client.Flush()

	// Should receive exactly 5 messages (one per message, not 15)
	count := 0
	timeout := time.After(5 * time.Second)
	for count < 5 {
		select {
		case <-received:
			count++
		case <-timeout:
			t.Fatalf("Timeout: received %d messages, expected 5", count)
		}
	}
}

func TestIntegration_JetStream_Stream(t *testing.T) {
	client, err := NewWithOptions(
		WithURL("nats://localhost:4222"),
		WithName("jetstream-test"),
		WithJetStream(true),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	streamName := "TEST_STREAM"

	// Cleanup from previous runs
	_ = client.DeleteStream(ctx, streamName)

	// Create stream
	streamConfig := &StreamConfig{
		Name:     streamName,
		Subjects: []string{"test.>"},
		Storage:  MemoryStorage,
		Replicas: 1,
	}

	info, err := client.CreateStream(ctx, streamConfig)
	if err != nil {
		t.Fatalf("CreateStream failed: %v", err)
	}
	t.Logf("Created stream: %s", info.Config.Name)

	// Get stream info
	info, err = client.GetStream(ctx, streamName)
	if err != nil {
		t.Fatalf("GetStream failed: %v", err)
	}
	if info.Config.Name != streamName {
		t.Errorf("Stream name mismatch: got %s, want %s", info.Config.Name, streamName)
	}

	// List streams
	streams, err := client.ListStreams(ctx)
	if err != nil {
		t.Fatalf("ListStreams failed: %v", err)
	}
	found := false
	for _, s := range streams {
		if s.Config.Name == streamName {
			found = true
			break
		}
	}
	if !found {
		t.Error("Stream not found in list")
	}

	// Publish to stream
	ack, err := client.PublishToStream(ctx, "test.events", []byte("event1"))
	if err != nil {
		t.Fatalf("PublishToStream failed: %v", err)
	}
	t.Logf("Published message: seq=%d, stream=%s", ack.Sequence, ack.Stream)

	// Get message
	msg, err := client.GetMsg(ctx, streamName, ack.Sequence)
	if err != nil {
		t.Fatalf("GetMsg failed: %v", err)
	}
	if string(msg.Data) != "event1" {
		t.Errorf("Message data mismatch: got %s, want event1", msg.Data)
	}

	// Purge stream
	if err := client.PurgeStream(ctx, streamName); err != nil {
		t.Fatalf("PurgeStream failed: %v", err)
	}

	// Delete stream
	if err := client.DeleteStream(ctx, streamName); err != nil {
		t.Fatalf("DeleteStream failed: %v", err)
	}
}

func TestIntegration_JetStream_Consumer(t *testing.T) {
	client, err := NewWithOptions(
		WithURL("nats://localhost:4222"),
		WithName("consumer-test"),
		WithJetStream(true),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	streamName := "CONSUMER_TEST"
	consumerName := "test-consumer"

	// Cleanup
	_ = client.DeleteStream(ctx, streamName)

	// Create stream
	_, err = client.CreateStream(ctx, &StreamConfig{
		Name:     streamName,
		Subjects: []string{"orders.>"},
		Storage:  MemoryStorage,
	})
	if err != nil {
		t.Fatalf("CreateStream failed: %v", err)
	}
	defer client.DeleteStream(ctx, streamName)

	// Create consumer
	consumerConfig := &ConsumerConfig{
		Durable:       consumerName,
		DeliverPolicy: DeliverAll,
		AckPolicy:     AckExplicit,
		AckWait:       30 * time.Second,
	}

	info, err := client.CreateConsumer(ctx, streamName, consumerConfig)
	if err != nil {
		t.Fatalf("CreateConsumer failed: %v", err)
	}
	t.Logf("Created consumer: %s", info.Name)

	// Get consumer info
	info, err = client.GetConsumer(ctx, streamName, consumerName)
	if err != nil {
		t.Fatalf("GetConsumer failed: %v", err)
	}
	if info.Name != consumerName {
		t.Errorf("Consumer name mismatch: got %s, want %s", info.Name, consumerName)
	}

	// List consumers
	consumers, err := client.ListConsumers(ctx, streamName)
	if err != nil {
		t.Fatalf("ListConsumers failed: %v", err)
	}
	if len(consumers) == 0 {
		t.Error("No consumers found")
	}

	// Delete consumer
	if err := client.DeleteConsumer(ctx, streamName, consumerName); err != nil {
		t.Fatalf("DeleteConsumer failed: %v", err)
	}
}

func TestIntegration_JetStream_PullSubscribe(t *testing.T) {
	client, err := NewWithOptions(
		WithURL("nats://localhost:4222"),
		WithName("pull-test"),
		WithJetStream(true),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	streamName := "PULL_TEST"

	// Cleanup
	_ = client.DeleteStream(ctx, streamName)

	// Create stream
	_, err = client.CreateStream(ctx, &StreamConfig{
		Name:     streamName,
		Subjects: []string{"tasks.>"},
		Storage:  MemoryStorage,
	})
	if err != nil {
		t.Fatalf("CreateStream failed: %v", err)
	}
	defer client.DeleteStream(ctx, streamName)

	// Publish messages
	for i := 0; i < 5; i++ {
		_, err := client.PublishToStream(ctx, "tasks.new", []byte("task"))
		if err != nil {
			t.Fatalf("Publish failed: %v", err)
		}
	}

	// Create pull subscription
	sub, err := client.PullSubscribe("tasks.>", "worker", nats.BindStream(streamName))
	if err != nil {
		t.Fatalf("PullSubscribe failed: %v", err)
	}
	defer sub.Unsubscribe()

	// Fetch messages
	msgs, err := sub.Fetch(5, nats.MaxWait(5*time.Second))
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(msgs) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(msgs))
	}

	// Ack all messages
	for _, msg := range msgs {
		msg.Ack()
	}
}

func TestIntegration_ChanSubscribe(t *testing.T) {
	client, err := NewWithOptions(
		WithURL("nats://localhost:4222"),
		WithName("chan-test"),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	subject := "test.chan"
	ch := make(chan *nats.Msg, 10)

	sub, err := client.ChanSubscribe(subject, ch)
	if err != nil {
		t.Fatalf("ChanSubscribe failed: %v", err)
	}
	defer sub.Unsubscribe()

	// Publish
	if err := client.Publish(subject, []byte("hello")); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Receive
	select {
	case msg := <-ch:
		if string(msg.Data) != "hello" {
			t.Errorf("Data mismatch: got %s, want hello", msg.Data)
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestIntegration_PublishMethods(t *testing.T) {
	client, err := NewWithOptions(
		WithURL("nats://localhost:4222"),
		WithName("publish-test"),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Test PublishMsg
	t.Run("PublishMsg", func(t *testing.T) {
		ch := make(chan *nats.Msg, 1)
		sub, _ := client.ChanSubscribe("test.publishmsg", ch)
		defer sub.Unsubscribe()

		msg := &nats.Msg{
			Subject: "test.publishmsg",
			Data:    []byte("test data"),
		}
		if err := client.PublishMsg(msg); err != nil {
			t.Fatalf("PublishMsg failed: %v", err)
		}

		select {
		case received := <-ch:
			if string(received.Data) != "test data" {
				t.Errorf("Data mismatch")
			}
		case <-time.After(5 * time.Second):
			t.Error("Timeout")
		}
	})

	// Test PublishRequest
	t.Run("PublishRequest", func(t *testing.T) {
		replySubject := "test.reply"
		ch := make(chan *nats.Msg, 1)
		sub, _ := client.ChanSubscribe(replySubject, ch)
		defer sub.Unsubscribe()

		// Setup responder
		responder, _ := client.Subscribe("test.publishrequest", func(msg *nats.Msg) {
			client.Publish(msg.Reply, []byte("response"))
		})
		defer responder.Unsubscribe()

		if err := client.PublishRequest("test.publishrequest", replySubject, []byte("request")); err != nil {
			t.Fatalf("PublishRequest failed: %v", err)
		}

		select {
		case received := <-ch:
			if string(received.Data) != "response" {
				t.Errorf("Data mismatch: got %s", received.Data)
			}
		case <-time.After(5 * time.Second):
			t.Error("Timeout")
		}
	})
}

func TestIntegration_RequestWithContext(t *testing.T) {
	client, err := NewWithOptions(
		WithURL("nats://localhost:4222"),
		WithName("ctx-request-test"),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	subject := "test.ctx.request"

	// Setup responder
	sub, _ := client.Subscribe(subject, func(msg *nats.Msg) {
		msg.Respond([]byte("ctx-response"))
	})
	defer sub.Unsubscribe()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reply, err := client.RequestWithContext(ctx, subject, []byte("ctx-request"))
	if err != nil {
		t.Fatalf("RequestWithContext failed: %v", err)
	}

	if string(reply.Data) != "ctx-response" {
		t.Errorf("Response mismatch: got %s", reply.Data)
	}
}
