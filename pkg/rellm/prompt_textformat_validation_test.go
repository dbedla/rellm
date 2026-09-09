package rellm_test

import (
	"testing"

	"rellm/pkg/rellm"

	"github.com/stretchr/testify/assert"
)

func TestPromptBuilder_Validation_RejectsNilTextFormat(t *testing.T) {
	_, err := rellm.NewPromptBuilder().
		WithMessage("Hi").
		WithTextFormat(nil).
		Build()

	assert.ErrorIs(t, err, rellm.ErrEmptyTextFormat)
}

func TestPromptBuilder_Validation_RejectsEmptyTextFormatType(t *testing.T) {
	_, err := rellm.NewPromptBuilder().
		WithMessage("Hi").
		WithTextFormat(&rellm.TextFormat{Name: "person"}).
		Build()

	assert.ErrorIs(t, err, rellm.ErrEmptyTextFormat)
}

func TestPromptBuilder_Validation_RejectsNilTextFormatSchema(t *testing.T) {
	_, err := rellm.NewPromptBuilder().
		WithMessage("Hi").
		WithTextFormat(&rellm.TextFormat{Type: "json_schema", Name: "person"}).
		Build()

	assert.ErrorIs(t, err, rellm.ErrEmptyTextFormat)
}
