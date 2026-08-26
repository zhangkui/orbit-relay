package service

import (
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/model"
)

func TestBug010_MissingWindowPreservesNotFoundAcrossLayers(t *testing.T) {
	lab := &Lab{windows: map[string]model.ContactWindow{}}
	if _, err := lab.GetWindow("missing-window"); err == nil {
		t.Fatal("missing window must not be converted into an empty successful result")
	}
}
