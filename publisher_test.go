package pubee_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/proto/proto3_proto"
	"github.com/google/go-cmp/cmp"
	"github.com/izumin5210/pubee"
)

type fakeDriver struct {
	Messages    []*pubee.Message
	PublishFunc func(context.Context, *pubee.Message)
}

var _ pubee.Driver = (*fakeDriver)(nil)

func (d *fakeDriver) Publish(ctx context.Context, msg *pubee.Message) {
	if f := d.PublishFunc; f != nil {
		f(ctx, msg)
		return
	}
	d.Messages = append(d.Messages, msg)
}
func (d *fakeDriver) Close(context.Context) error { return nil }

func TestPublisher_WithMetadata(t *testing.T) {
	driver := new(fakeDriver)
	publisher := pubee.NewPublisher(driver,
		pubee.WithMetadataMap(map[string]string{"foo": "1", "bar": "2"}),
	)
	err := publisher.Publish(
		context.Background(),
		"foobarbaz",
		pubee.WithMetadata("baz", "3", "foo", "foooooo"),
	)
	if err != nil {
		t.Errorf("Publish() returns %v, want nil", err)
	}
	if got, want := len(driver.Messages), 1; got != want {
		t.Errorf("Published messages are %d, want %d", got, want)
	} else {
		msg := driver.Messages[0]
		if got, want := string(msg.Data), "foobarbaz"; got != want {
			t.Errorf("Publish message has data %v, want %v", got, want)
		}
		md := map[string]string{"foo": "foooooo", "bar": "2", "baz": "3"}
		if got, want := msg.Metadata, md; !reflect.DeepEqual(got, want) {
			t.Errorf("Publish message has metadata %v, want %v", got, want)
		}
	}
}

func TestPublisher_WithInterceptors(t *testing.T) {
	var ops []string
	driver := new(fakeDriver)
	publisher := pubee.NewPublisher(driver,
		pubee.WithInterceptors(
			func(ctx context.Context, msg *pubee.Message, handle func(context.Context, *pubee.Message)) {
				ops = append(ops, "1-before")
				handle(ctx, msg)
				ops = append(ops, "1-after")
			},
			func(ctx context.Context, msg *pubee.Message, handle func(context.Context, *pubee.Message)) {
				ops = append(ops, "2-before")
				handle(ctx, msg)
				ops = append(ops, "2-after")
			},
			func(ctx context.Context, msg *pubee.Message, handle func(context.Context, *pubee.Message)) {
				ops = append(ops, "3-before")
				handle(ctx, msg)
				ops = append(ops, "3-after")
			},
		),
	)
	err := publisher.Publish(
		context.Background(),
		"foobarbaz",
		pubee.WithMetadata("baz", "3", "foo", "foooooo"),
	)
	if err != nil {
		t.Errorf("Publish() returns %v, want nil", err)
	}
	if got, want := len(driver.Messages), 1; got != want {
		t.Errorf("Published messages are %d, want %d", got, want)
	}
	if got, want := ops, []string{
		"1-before", "2-before", "3-before",
		"3-after", "2-after", "1-after",
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("interceptors called order is %v, want %v", got, want)
	}
}

func TestPublisher_WithJSON(t *testing.T) {
	driver := new(fakeDriver)
	publisher := pubee.NewPublisher(driver, pubee.WithJSON())
	err := publisher.Publish(
		context.Background(),
		"foobarbaz",
	)
	if err != nil {
		t.Errorf("Publish() returns %v, want nil", err)
	}
	if got, want := len(driver.Messages), 1; got != want {
		t.Errorf("Published messages are %d, want %d", got, want)
	} else {
		msg := driver.Messages[0]
		if got, want := string(msg.Data), `"foobarbaz"`; got != want {
			t.Errorf("Publish message has data %v, want %v", got, want)
		}
	}
}

func TestPublisher_WithProtobuf(t *testing.T) {
	driver := new(fakeDriver)
	publisher := pubee.NewPublisher(driver, pubee.WithProtobuf())
	in := &proto3_proto.Message{Name: "Foo Bar", Hilarity: proto3_proto.Message_PUNS}
	err := publisher.Publish(context.Background(), in)
	if err != nil {
		t.Errorf("Publish() returns %v, want nil", err)
	}
	if got, want := len(driver.Messages), 1; got != want {
		t.Errorf("Published messages are %d, want %d", got, want)
	} else {
		msg := driver.Messages[0]
		var out proto3_proto.Message
		err = proto.Unmarshal(msg.Data, &out)
		if err != nil {
			t.Errorf("failed to unmarshal publishhed message: %v", err)
		}
		if diff := cmp.Diff(*in, out, cmp.FilterPath(func(p cmp.Path) bool {
			return strings.HasPrefix(p.Last().String(), ".XXX_")
		}, cmp.Ignore())); diff != "" {
			t.Errorf("Publish message mismatch(-want, +got):\n%s", diff)
		}
	}
}

func TestPublisher_WithOnFailPublish(t *testing.T) {
	driver := &fakeDriver{
		PublishFunc: func(ctx context.Context, msg *pubee.Message) {
			pubee.GetDriverCallback(ctx).OnFailPublish(msg, errors.New("unfortunate error"))
		},
	}
	var calledCnt int
	publisher := pubee.NewPublisher(driver,
		pubee.WithOnFailPublish(func(msg *pubee.Message, err error) {
			calledCnt++
			if got, want := string(msg.Data), "foobarbaz"; got != want {
				t.Errorf("OnFailPublish received message %q, want %q", got, want)
			}
			if got, want := err.Error(), "unfortunate error"; got != want {
				t.Errorf("OnFailPublish received error %q, want %q", got, want)
			}
		}),
	)
	err := publisher.Publish(context.Background(), "foobarbaz")
	if err != nil {
		t.Errorf("Publish() returns %v, want nil", err)
	}
	if got, want := calledCnt, 1; got != want {
		t.Errorf("OnFailPublish is called %d times, want %d", got, want)
	}
}
