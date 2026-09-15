package main

import (
	"context"
	"fmt"
	"rellm/pkg/rellm"
	"strings"

	"github.com/fatih/color"
)

// Scenario reproduces the provider-switch test: one shared conversation is
// used first by an OpenRouter (luna) agent, then by an LMStudio agent. The
// conversation elements produced by the first provider must be loadable and
// serializable by the second.
func main() {
	conversation := rellm.NewInMemoryConversation()

	orAgent, err := buildORLunaSwitchAgent(conversation)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	beforeFirst, err := orAgent.Conversation().Load(ctx)
	if err != nil {
		panicWithLog("cannot load conversation before first ask", err)
	}
	if len(beforeFirst) != 0 {
		panicWithLog("expected empty conversation before first ask", fmt.Errorf("conversation has %d elements", len(beforeFirst)))
	}

	msg := "hi"
	color.Magenta(msg)

	firstResp, err := orAgent.Ask(ctx, msg)
	if err != nil {
		panicWithLog("openrouter agent failed for input msg: "+msg, err)
	}
	color.Green(firstResp.Message)

	afterFirst, err := orAgent.Conversation().Load(ctx)
	if err != nil {
		panicWithLog("cannot load conversation after openrouter turn", err)
	}
	if len(afterFirst) < 2 {
		panicWithLog("expected conversation elements from the openrouter turn", fmt.Errorf("conversation has %d elements", len(afterFirst)))
	}

	msg = "what tools do you see?"
	color.Magenta(msg)

	secondResp, err := orAgent.Execute(ctx, mustPrompt(msg))
	if err != nil {
		panicWithLog("lms agent failed with an openrouter-built conversation, msg: "+msg, err)
	}
	color.Green(secondResp.Message)

	if !strings.Contains(secondResp.Message, "GetDataFor") || !strings.Contains(secondResp.Message, "GetStaticData") {
		panicWithLog("missing expected tool names in lms output", fmt.Errorf("unexpected message"))
	}

	//LMS take over
	lmsAgent, err := buildLMSSwitchAgent(conversation)
	if err != nil {
		panic(err)
	}

	afterSecond, err := lmsAgent.Conversation().Load(ctx)
	if err != nil {
		panicWithLog("cannot load conversation after lms turn", err)
	}
	if len(afterSecond) <= len(afterFirst) {
		panicWithLog("expected conversation to grow after lms turn", fmt.Errorf("conversation has %d elements, was %d", len(afterSecond), len(afterFirst)))
	}
}

func mustPrompt(msg string) *rellm.Prompt {
	prompt, err := rellm.NewPromptBuilder().
		WithMessage(msg).
		WithReasoning(rellm.ReasoningEffortLow).
		Build()
	if err != nil {
		panicWithLog("unable to build prompt", err)
	}
	return prompt
}

func panicWithLog(msg string, err error) {
	logMsg := msg + "\n" + err.Error()
	color.Red(logMsg)
	panic(logMsg)
}
