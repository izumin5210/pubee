package cloudpubsub_test

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"github.com/izumin5210/pubee"
	"github.com/izumin5210/pubee/drivers/cloudpubsub"
)

type pubsubtest struct {
	Server *pstest.Server
}

func (pst *pubsubtest) Conn(t *testing.T) *grpc.ClientConn {
	t.Helper()

	conn, err := grpc.Dial(pst.Server.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to connect pstest.Server: %v", err)
	}

	return conn
}

func (pst *pubsubtest) Client(t *testing.T) *pubsub.Client {
	t.Helper()

	client, err := pubsub.NewClient(context.Background(), "awesomeproj", option.WithGRPCConn(pst.Conn(t)))
	if err != nil {
		t.Fatalf("failed to create pubsub.Client: %v", err)
	}

	return client
}

func (pst *pubsubtest) TopicExists(t *testing.T, id string) bool {
	t.Helper()

	client := pst.Client(t)
	defer client.Close()

	ok, err := client.Topic("awesometopic").Exists(context.Background())
	if err != nil {
		t.Fatalf("failed to check topic existence: %v", err)
	}

	return ok
}

func (pst *pubsubtest) Close() {
	if pst.Server != nil {
		pst.Server.Close()
	}
}

func newPubsubTest(t *testing.T) *pubsubtest {
	t.Helper()

	srv := pstest.NewServer()

	return &pubsubtest{
		Server: srv,
	}
}

func TestDriver(t *testing.T) {
	pst := newPubsubTest(t)
	defer pst.Close()

	ctx := context.Background()

	_, err := pst.Client(t).CreateTopic(ctx, "awesometopic")
	if err != nil {
		t.Fatalf("failed to create pubsub.Topic: %v", err)
	}

	driver, err := cloudpubsub.CreateDriver(ctx,
		"awesomeproj",
		"awesometopic",
		cloudpubsub.WithClientOptions(option.WithGRPCConn(pst.Conn(t))),
	)
	if err != nil {
		t.Fatalf("failed to create a cloudpubsub.Driver: %v", err)
	}

	err = <-driver.Publish(ctx, &pubee.Message{Data: []byte("test message")})
	if err != nil {
		t.Errorf("failed to publish a message: %v", err)
	}
	driver.Close(ctx)

	if got, want := len(pst.Server.Messages()), 1; got != want {
		t.Errorf("Received messages are %d, want %d", got, want)
	}
}

func TestDriver_WithoutTopic(t *testing.T) {
	pst := newPubsubTest(t)
	defer pst.Close()

	ctx := context.Background()

	_, err := cloudpubsub.CreateDriver(ctx,
		"awesomeproj",
		"awesometopic",
		cloudpubsub.WithClientOptions(option.WithGRPCConn(pst.Conn(t))),
	)
	if err == nil {
		t.Error("CreateDriver should return an error")
	}
}

func TestDriver_WithCreateTopic(t *testing.T) {
	pst := newPubsubTest(t)
	defer pst.Close()

	ctx := context.Background()

	driver, err := cloudpubsub.CreateDriver(ctx,
		"awesomeproj",
		"awesometopic",
		cloudpubsub.WithClientOptions(option.WithGRPCConn(pst.Conn(t))),
		cloudpubsub.WithCreateTopicIfNeeded(),
	)
	if err != nil {
		t.Fatalf("failed to create a cloudpubsub.Driver: %v", err)
	}

	if got, want := pst.TopicExists(t, "awesometopic"), true; got != want {
		t.Errorf("Topic.Exists() returned %t, want %t", got, want)
	}

	err = driver.Close(ctx)
	if err != nil {
		t.Errorf("Driver.Close() returned %v, want %v", err, nil)
	}

	if got, want := pst.TopicExists(t, "awesometopic"), true; got != want {
		t.Errorf("Topic.Exists() returned %t, want %t", got, want)
	}
}

func TestDriver_WithCreateTopic_and_WithDeleteTopic(t *testing.T) {
	pst := newPubsubTest(t)
	defer pst.Close()

	ctx := context.Background()

	driver, err := cloudpubsub.CreateDriver(ctx,
		"awesomeproj",
		"awesometopic",
		cloudpubsub.WithClientOptions(option.WithGRPCConn(pst.Conn(t))),
		cloudpubsub.WithCreateTopicIfNeeded(),
		cloudpubsub.WithDeleteTopicOnClose(),
	)
	if err != nil {
		t.Fatalf("failed to create a cloudpubsub.Driver: %v", err)
	}

	if got, want := pst.TopicExists(t, "awesometopic"), true; got != want {
		t.Errorf("Topic.Exists() returned %t, want %t", got, want)
	}

	err = driver.Close(ctx)
	if err != nil {
		t.Errorf("Driver.Close() returned %v, want %v", err, nil)
	}

	if got, want := pst.TopicExists(t, "awesometopic"), false; got != want {
		t.Errorf("Topic.Exists() returned %t, want %t", got, want)
	}
}

type fakeDriverCallback struct {
	OnFailPublishFunc func(*pubee.Message, error)
}

func (cb *fakeDriverCallback) OnFailPublish(msg *pubee.Message, err error) {
	cb.OnFailPublishFunc(msg, err)
}

func TestDriver_OnFailPublishCalled(t *testing.T) {
	pst := newPubsubTest(t)
	defer pst.Close()

	ctx := context.Background()

	driver, err := cloudpubsub.CreateDriver(ctx,
		"awesomeproj",
		"awesometopic",
		cloudpubsub.WithClientOptions(option.WithGRPCConn(pst.Conn(t))),
		cloudpubsub.WithCreateTopicIfNeeded(),
		cloudpubsub.WithPublishSettings(func(s *pubsub.PublishSettings) {
			s.CountThreshold = 1
			s.Timeout = 30 * time.Millisecond
		}),
	)
	if err != nil {
		t.Fatalf("failed to create a cloudpubsub.Driver: %v", err)
	}

	var calledCnt int

	errChCh := make(chan (<-chan error), 3)

	errChCh <- driver.Publish(ctx, &pubee.Message{Data: []byte("test message1")})
	time.Sleep(30 * time.Millisecond)
	pst.Server.Close()
	errChCh <- driver.Publish(ctx, &pubee.Message{Data: []byte("test message2")})
	errChCh <- driver.Publish(ctx, &pubee.Message{Data: []byte("test message3")})
	driver.Close(ctx)
	close(errChCh)

	for errCh := range errChCh {
		for err := range errCh {
			if err != nil {
				calledCnt++
			}
		}
	}

	if got, want := calledCnt, 2; got != want {
		t.Errorf("OnFailPublish is called %d times, want %d", got, want)
	}
}
