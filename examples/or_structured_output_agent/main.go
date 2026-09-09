package main

import (
	"context"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"

	"github.com/fatih/color"
)

func main() {
	agent, err := buildStructuredOutputAgent()
	if err != nil {
		panic(err)
	}

	textFormat := &rellm.TextFormat{
		Type:   "json_schema",
		Name:   "person",
		Strict: true,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{"type": "string"},
				"age":  map[string]interface{}{"type": "number"},
				"city": map[string]interface{}{"type": "string"},
			},
			"required":             []string{"name", "age", "city"},
			"additionalProperties": false,
		},
	}

	for {
		msg, err := examplesutils.ReadConsoleInput()
		if err != nil {
			color.Red("unable to read console input: %s", err)
			return
		}
		if msg == "EXIT" {
			return
		}

		prompt, err := rellm.NewPromptBuilder().
			WithMessage(msg).
			WithTextFormat(textFormat).
			Build()
		if err != nil {
			color.Red("unable to build prompt: %s", err.Error())
			continue
		}

		ctx := context.Background()
		llmResp, err := agent.Execute(ctx, prompt)
		if err != nil {
			color.Red("unable to ask question: %s", err.Error())
			continue
		}

		color.Blue(llmResp)
	}
}
