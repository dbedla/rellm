package main

import (
	"context"
	"encoding/json"
	"rellm/pkg/rellm"

	"github.com/fatih/color"
	"github.com/invopop/jsonschema"
)

var textFormat = &rellm.TextFormat{
	Type:   "json_schema",
	Name:   "person",
	Strict: true,
	Schema: (&jsonschema.Reflector{DoNotReference: true}).Reflect(&Person{}),
}

// Person is the structured shape we ask the model to produce.
type Person struct {
	Name string `json:"name" jsonschema:"description=Full name of the person"`
	Age  int    `json:"age"  jsonschema:"description=Age in years"`
	City string `json:"city" jsonschema:"description=City of residence"`
}

func main() {
	lmsAgent, err := buildLMSStructuredOutputAgent()
	if err != nil {
		panic(err)
	}
	orGeminiAgent, err := buildORGeminiStructuredOutputAgent()
	if err != nil {
		panic(err)
	}
	orLunaAgent, err := buildORLunaStructuredOutputAgent()
	if err != nil {
		panic(err)
	}
	agents := []*rellm.Agent{lmsAgent, orGeminiAgent, orLunaAgent}

	for _, a := range agents {
		scenario(a)
	}
}

func scenario(agent *rellm.Agent) {

	msg := "I am John Snow from Winterfell, I have 100 years..."
	prompt, err := rellm.NewPromptBuilder().
		WithMessage(msg).
		WithTextFormat(textFormat).
		Build()
	if err != nil {
		color.Red("unable to build prompt: %s", err.Error())
		panic(err)
	}

	ctx := context.Background()
	llmResp, err := agent.Execute(ctx, prompt)
	if err != nil {
		color.Red("unable to ask question: %s", err.Error())
		panic(err)
	}

	color.Blue("raw string output for agent: %s", agent.Name())
	color.Blue(llmResp.Messages)

	var person Person
	err = json.Unmarshal([]byte(llmResp.Messages), &person)
	if err != nil {
		color.Red("unable to parse json: %s", err.Error())
		panic(err)
	}
	color.Blue("after unmarshall to struct: %+v", person)
	if person.Name != "John Snow" || person.Age != 100 || person.City != "Winterfell" {
		panic("invalid data in struct")
	}
}
