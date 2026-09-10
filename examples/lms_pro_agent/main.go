package main

import (
	"context"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"

	"github.com/fatih/color"
)

func main() {
	agent, err := buildProAgent()
	if err != nil {
		panic(err)
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
			WithReasoning(rellm.ReasoningEffortLow).
			WithTemperature(0.5).
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
		color.Blue(llmResp.Messages)
		//if rawRsp != nil {
		//	color.Yellow("This conversation total cost in tokens: %d\n", rawRsp.Usage.TotalTokens)
		//	color.Yellow("input tokens: %d\n", rawRsp.Usage.InputTokens)
		//	color.Yellow("output tokens: %d\n", rawRsp.Usage.OutputTokens)
		//}
	}
}
