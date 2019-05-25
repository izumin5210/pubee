package pubee

type PublishConfig struct {
	Metadata map[string]string
}

type PublishOption func(*PublishConfig)
