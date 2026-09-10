package main

import (
	"context"
	"fmt"
	"os"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"

	"github.com/fatih/color"
)

func main() {

	if len(os.Args) < 2 {
		color.Red("please provide a provider option: --or-luna, --or-gemma, --or-gemini, --lms")
		return
	}

	agent, err := agentForProvider(os.Args[1])
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

func agentForProvider(provider string) (*rellm.Agent, error) {
	switch provider {
	case "--or-luna":
		return buildOpenRouterAgent("openai/gpt-5.6-luna")
	case "--or-gemma":
		return buildOpenRouterAgent("google/gemma-4-26b-a4b-it")
	case "--or-gemini":
		return buildOpenRouterAgent("google/gemini-3.1-flash-lite")
	case "--lms":
		return buildLMSAgent()
	default:
		return nil, fmt.Errorf("provider %s not supported, available options: --or-luna, --or-gemma, --or-gemini, --lms", provider)
	}
}
