package main

import (
	"context"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"

	"github.com/fatih/color"
	"github.com/invopop/jsonschema"
)

// Person is the structured shape we ask the model to produce.
type Person struct {
	Name string `json:"name" jsonschema:"description=Full name of the person"`
	Age  int    `json:"age"  jsonschema:"description=Age in years"`
	City string `json:"city" jsonschema:"description=City of residence"`
}

func main() {
	agent, err := buildStructuredOutputAgent()
	if err != nil {
		panic(err)
	}

	textFormat := &rellm.TextFormat{
		Type:   "json_schema",
		Name:   "person",
		Strict: true,
		Schema: (&jsonschema.Reflector{DoNotReference: true}).Reflect(&Person{}),
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
