package main

import (
	"context"
	"fmt"
	"os"
	"rellm/pkg/agentsutils"

	"rellm/pkg/rellm"

	"github.com/fatih/color"
)

func main() {

	if len(os.Args) < 2 {
		color.Red("please provide a provider option: --or-gemma, --or-gemini, --lms")
		return
	}

	agent, err := agentForProvider(os.Args[1])
	if err != nil {
		panic(err)
	}

	for {
		msg, err := agentsutils.ReadConsoleInput()
		if err != nil {
			color.Red("unable to read console input: %s", err)
			return
		}
		if msg == "EXIT" {
			return
		}

		prompt, err := rellm.NewPromptBuilder().
			WithMessage(msg).
			WithReasoning(rellm.ReasoningEffort_Low).
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
		color.Blue(llmResp)
		//if rawRsp != nil {
		//	color.Yellow("This conversation total cost in tokens: %d\n", rawRsp.Usage.TotalTokens)
		//	color.Yellow("input tokens: %d\n", rawRsp.Usage.InputTokens)
		//	color.Yellow("output tokens: %d\n", rawRsp.Usage.OutputTokens)
		//}
	}
}

func agentForProvider(provider string) (*rellm.Agent, error) {
	switch provider {
	case "--or-gemma":
		return buildLOpenRouterAgent(rellm.Model_OpenRouter_Google_Gemma_4_26b_A4b_It)
	case "--or-gemini":
		return buildLOpenRouterAgent(rellm.Model_OpenRouter_Google_Gemini_3_1_Flash_Lite)
	case "--lms":
		return buildLMSAgent()
	default:
		return nil, fmt.Errorf("provider %s not supported, available options: --or-gemma, --or-gemini, --lms", provider)
	}
}
