package pubee

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	defaultErrorLog = new(nopLogger)
	os.Exit(m.Run())
}
