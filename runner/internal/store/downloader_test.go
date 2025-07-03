package store

import (
	"context"
	"testing"
)

func TestHFModelInfo(t *testing.T) {
	s := Get()
	s.HFModelInfo(context.Background(), "nexaml/nexaml-models")
}
