// Package rellm is an experimental Go library for building custom agents based
// on the Responses API endpoint (as proposed by OpenAI).
//
// The Responses API is implemented by more than OpenAI: LM Studio serves it for
// locally run models, and OpenRouter exposes it as a unified gateway to many
// hosted models. rellm can talk to either one.
//
// # Quick start
//
// Build a provider, wire up an agent, and ask a question:
//
//	provider, err := rellm.NewLMStudioProvider(
//		"google/gemma-4-26b-a4b",
//		"http://127.0.0.1",
//		"1234",
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	agent, err := rellm.NewAgentBuilder().
//		WithProvider(provider).
//		WithConversationStorage(rellm.NewInMemoryStorage()).
//		WithUnknownConversationElementKeepInTheLoop().
//		WithImageGenerationKeepInTheLoop().
//		Build()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	finalReport, err := agent.Ask(context.Background(), "What is the meaning of life?")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(finalReport.Messages)
//
// The Agent owns an agentic loop: it manages conversation history, assembles
// the HTTP requests, and dispatches tool (function) calls through a Toolset
// until the model produces a final answer. For a richer Prompt with extra
// inference parameters, use PromptBuilder and Agent.Execute.
//
// See the examples/ and e2e_tests/ directories in the repository for complete
// programs, including an agent extended with filesystem tools.
package rellm
