# pubee

[![GoDoc](https://godoc.org/github.com/izumin5210/pubee?status.svg)](https://godoc.org/github.com/izumin5210/pubee)

## Example

```go
ctx := context.Background()

// Initiailze a driver for Google Cloud Pub/Sub
driver, err := cloudpubsub.NewDriver(
	ctx, "my-gcp-project", "your-topic",
	cloudpubsub.WithCreateTopicIfNeeded(),  // Create a topic when it does not exist
	cloudpubsub.WithDeleteTopicOnClose(),   // Delete the topic on close the publisher
)
if err != nil {
	// ...
}

// Initialize a new publisher instance
publisher := pubee.NewPublisher(driver,
	pubee.WithJSON(),  // publish messages as JSON
	pubee.WithMetadata("content_type", "json"),
	pubee.WithInterceptors(
		// ...
	}),
	pubee.WithOnFailPublish(func(msg *pubee.Message, err error {
		// ...
	}),
)
defer publisher.Close()

type Book struct {
	Title  string `json:"title"`
}

// Publish a message!
err := publisher.Publish(ctx, &Book{Title: "The Go Programming Language"})
if err != nil {
	// ...
}
```
