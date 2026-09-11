package rellm_test

import (
	"testing"

	"rellm/pkg/rellm"

	"github.com/stretchr/testify/assert"
)

func TestAgentBuilder_Validation_RejectsEmptyTextFormatType(t *testing.T) {
	_, err := rellm.NewAgentBuilder().
		WithTextFormat(&rellm.TextFormat{Name: "person"}).
		Build()

	assert.ErrorIs(t, err, rellm.ErrEmptyTextFormat)
}

func TestAgentBuilder_Validation_RejectsNilTextFormatSchema(t *testing.T) {
	_, err := rellm.NewAgentBuilder().
		WithTextFormat(&rellm.TextFormat{Type: "json_schema", Name: "person"}).
		Build()

	assert.ErrorIs(t, err, rellm.ErrEmptyTextFormat)
}
