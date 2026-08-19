package main

import (
	"rellm/pkg/agentsutils"
	"rellm/pkg/rellm"

	"github.com/fatih/color"
)

func main() {
	agentLogDir, err := agentsutils.CreateDirInSysTmp("agent-log")
	if err != nil {
		panic(err)
	}

	color.Red("Agent log dir: %s", agentLogDir)

	agent, err := buildLBaseAgent(agentLogDir)
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
			WithReasoning(rellm.ReasoningEffort_Medium).
			Build()

		if err != nil {
			color.Red("unable to build prompt: %s", err.Error())
			continue
		}

		llmResp, err := agent.Execute(prompt)
		if err != nil {
			color.Red("unable to ask question: %s", err.Error())
			continue
		}
		//if rawResp != nil {
		//	color.Yellow("This conversation total cost in tokens: %d\n", rawResp.Usage.TotalTokens)
		//	if rawResp.Usage.Cost != nil {
		//		color.Yellow("This conversation total cost in USD: %f\n", *rawResp.Usage.Cost)
		//	}
		//}
		color.Blue(llmResp)
	}

}
