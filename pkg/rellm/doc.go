// Package rellm is a lightweight Go framework for building custom agents on
// top of the Responses API (/v1/responses). The endpoint is implemented by
// more than OpenAI: LM Studio serves it for locally run models, and
// OpenRouter exposes it as a unified gateway to many hosted models, so local
// and hosted models can be used interchangeably.
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
//		WithConversation(rellm.NewInMemoryConversation()).
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
//	fmt.Println(finalReport.Message)
//
// The Agent owns an agentic loop: it manages conversation history, assembles
// the HTTP requests, and dispatches tool (function) calls through a Toolset
// until the model produces a final answer. For a richer Prompt with extra
// inference parameters, use PromptBuilder and Agent.Execute.
//
// # Tools
//
// Tools define how the agent can interact with your system — you decide
// exactly what the agent is allowed to do, and fewer tools mean fewer ways
// for a hallucination to cause side effects. Implement the Toolset interface
// and register it with WithToolset on the builder:
//
//	type Toolset interface {
//		// Definitions returns the tool definitions advertised to the model.
//		Definitions() []rellm.ToolDefinition
//
//		// Dispatch executes a tool call requested by the model.
//		Dispatch(ctx context.Context, name string, arguments json.RawMessage) (rellm.ToolCallResult, error)
//	}
//
// Keep functionality separate from the toolset: write the logic as a plain
// struct with methods, then wrap it in a thin Toolset adapter that only
// translates between the model and your code. This way the functionality is
// testable as ordinary Go code, with no model or agent in the loop. See the
// Calculator sample in pkg/toolsets for this pattern in full.
//
// # Structured output (JSON schema)
//
// Instead of parsing free text, the agent can return a schema-constrained
// JSON structure. Generate the schema from a Go struct with a reflector
// (for example github.com/invopop/jsonschema), describe each field with a
// jsonschema struct tag, and pass it via WithTextFormat:
//
//	type Person struct {
//		Name string `json:"name" jsonschema:"description=Full name of the person"`
//		Age  int    `json:"age"  jsonschema:"description=Age in years"`
//	}
//
//	textFormat := rellm.TextFormat{
//		Type:   "json_schema",
//		Name:   "person",
//		Strict: true,
//		Schema: (&jsonschema.Reflector{DoNotReference: true}).Reflect(&Person{}),
//	}
//
//	agent, err := rellm.NewAgentBuilder().
//		...
//		WithTextFormat(textFormat).
//		Build()
//
// A JSON answer is far easier to put to work than free text: unmarshal it,
// pass it to another tool, or persist it without fragile string scraping —
// every field carries a type and a description the model generally respects.
// A narrow toolset and a strict output schema constrain the model at both
// ends: what it can do and what it can say.
//
// # Image inputs
//
// Vision prompts attach images to the user message as input_image parts,
// built with the PromptBuilder's WithImage* methods. Every method takes an
// ImageOption as its last parameter; its zero value means "all defaults"
// (detail defaults to auto):
//
//	prompt, err := rellm.NewPromptBuilder().
//		WithMessage("what is in this image?").
//		WithImageURL("https://example.com/cat.png", rellm.ImageOption{}).
//		WithImageDataURL("data:image/png;base64,...", rellm.ImageOption{}).
//		WithImageBytes(pngBytes, rellm.ImageOption{}).
//		WithImageFile("/tmp/cat.png", rellm.ImageOption{}).
//		WithImageFileID("file-123", rellm.ImageOption{}).
//		Build()
//
// WithImageURL sends a web URL the provider fetches (data URLs are rejected
// there; use WithImageDataURL). WithImageBytes and WithImageFile sniff the
// MIME type and base64-encode the data into a data URL at Build.
// WithImageFileID references an image uploaded via the provider's Files API.
// The ImageOption of every method carries the doc-mentioned optional fields:
// Detail (defaults to auto), FileID (WithImageFileID only; mutually exclusive
// with the image URL), and CacheBreakpoint (an OpenAI-style prompt-cache
// boundary; OpenRouter converts it for Anthropic-style providers). A message
// or at least one image is required; repeat calls accumulate images in call
// order. Data-URL images persist into the conversation history (JSONL) and
// are re-sent on every round — host the image, upload it, or pass a URL to
// keep history small.
//
// # Examples
//
// The examples/ and e2e_tests/ directories in the repository contain
// complete runnable programs, including agents extended with filesystem and
// calculator toolsets.
//
// For a real-world application built with rellm, see cluesh, an example of
// usage of this framework.
//
// # Links
//
// - Homepage, https://rellm.dev/
// - GitHub repository, https://github.com/dbedla/rellm
// - cluesh (example usage), https://rellm.dev/agents/cluesh
// - cluesh source, https://github.com/dbedla/cluesh
package rellm
