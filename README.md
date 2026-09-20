# rellm

`rellm` (from **Res**ponses API + **LLM** + endpoint) is a lightweight Go
framework for building custom agents on top of the OpenAI **Responses API**
(`/v1/responses`). Multiple providers implement this endpoint (LM Studio,
OpenRouter), which lets you freely switch between local and hosted models
to balance cost and result quality.

A key feature is easy custom toolset injection: you decide exactly how the
agent can interact with your system. More control over the toolset means
fewer unexpected side effects when the model hallucinates (see
[Adding tools to your agent](#adding-tools-to-your-agent)).

Another is a predefined JSON output format: instead of parsing free text, the
agent returns a schema-validated structure (see
[Structured output (JSON schema)](#structured-output-json-schema)). Combine
the two — a narrow toolset and a strict output schema — and a small,
specialized agent becomes very powerful: the model is constrained at both
ends, on what it can do and on what it can say.

## Installation
```bash
go get github.com/dbedla/rellm/pkg/rellm
```

## Table of contents

- [Why small, specialized agents?](#why-small-specialized-agents)
- [Supported providers & models](#supported-providers--models)
- [Quick start](#quick-start)
  - [Run the example agent](#run-the-example-agent)
  - [Build your own agent](#build-your-own-agent)
  - [Structured output (JSON schema)](#structured-output-json-schema)
  - [Adding tools to your agent](#adding-tools-to-your-agent)
- [Tips](#tips)
- [External links](#external-links)

## Why small, specialized agents?

`rellm` is built around the idea that many small, focused agents beat one
general-purpose one:

- **Fewer tools, fewer mistakes.** A small toolset gives the model fewer
  ways to pick the wrong tool, so hallucinations have a smaller blast radius.
- **Cheaper and faster.** A narrow task needs less model capability — small
  local models often do the job, so you avoid paying for a frontier model.
- **Easier to test and measure.** One agent, one responsibility: success
  criteria are clear, and regressions are easy to spot.
- **Composable.** Each agent is a plain Go package — chain them, run them in
  parallel, or embed one in another agent's toolset.
- **Fits workflows naturally.** A small agent is a single step: run it as part
  of a pipeline, embed it in a larger system, or call it from ordinary Go code
  wherever a decision or an answer is needed.
- **Great for categorization.** Classification, routing, and tagging tasks have
  a bounded output and a measurable accuracy — ideal for a small agent running
  on a cheap model. A categorization agent can also route work to the right,
  more specialized agent.

## Supported providers & models

`rellm` talks to providers through a provider interface, so any backend that
implements the Responses API can be used.

Officially supported providers:

- **LM Studio** — run models locally
- **OpenRouter** — unified gateway to many hosted models
- **OpenAI** — direct access to OpenAI models

Need something else? The provider interface is public — implement your own
as needed.

Recommended local models (via LM Studio):

- `deepreinforce-ai/ornith-1.5-35b-a3b` (older: `deepreinforce-ai/ornith-1.0-35b`)
- `google/gemma-4-26b-a4b`
- `qwen/qwen3.8-27b`
- `meta/muse-glimmer`

Benchmarks:

- [OpenRouter benchmarks](https://openrouter.ai/benchmarks)
- [LLM Stats (SWE-bench Verified)](https://llm-stats.com/benchmarks/swe-bench-verified)
- [SWE-rebench](https://swe-rebench.com)
- [Berkeley Function-Calling Leaderboard](https://gorilla.cs.berkeley.edu/leaderboard.html)
- [Artificial Analysis](https://artificialanalysis.ai)
- [LiveBench](https://livebench.ai)
- [LMArena](https://arena.ai/leaderboard)
- [Model Evaluation & Threat Research](https://metr.org)

## Quick start

### Run the example agent

Requires [LM Studio](https://lmstudio.ai/download) with its [CLI](https://lmstudio.ai/docs/cli).

From the repository root:

```bash
# download and start the local model (gemma-4-26b-a4b)
make e2e-lms-env

# build and run the example agent
make go-build
./output/bin/endpoint_agent --lms
```

Type a message to chat with the agent. Type `EXIT` to quit.

To use OpenRouter instead, set up a `.env` file with an API key
([generate one here](https://openrouter.ai/workspaces/default/keys)):

```
OPENROUTER_API_KEY=sk-...
```

then run with a provider flag, e.g. `./output/bin/endpoint_agent --or-gemma`.
Other provider flags: `--lms`, `--or-gemma`, `--or-gemini`, `--or-luna`.

To use OpenAI directly, set up a `.env` file with an API key
([generate one here](https://platform.openai.com/api-keys)):

```
OPENAI_API_KEY=sk-...
```

The OpenAI provider is used by the `e2e_tests` binaries via the `--openai`
flag, e.g. `./output/bin/e2e_fs_agent --openai` (model `gpt-5.6-luna`).

### Build your own agent

```go
package main

import (
	"context"
	"fmt"

	"rellm/pkg/rellm"
)

func main() {
	provider, err := rellm.NewLMStudioProvider(
		"google/gemma-4-26b-a4b", "http://127.0.0.1", "1234",
	)
	if err != nil {
		panic(err)
	}

	agent, err := rellm.NewAgentBuilder().
		WithProvider(provider).
		WithAgentName("MyAgent").
		WithMaxAgentSteps(20).
		WithConversation(rellm.NewInMemoryConversation()).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		WithSystemMessage("You are a helpful assistant.").
		Build()
	if err != nil {
		panic(err)
	}

	finalReport, err := agent.Ask(context.Background(), "Hello!")
	if err != nil {
		panic(err)
	}
	fmt.Println(finalReport.Message)
}
```

See the full runnable version in [examples/endpoint_agent](examples/endpoint_agent),
which shows how to switch between providers with flags.
Other examples:

- [examples/lms_base_agent](examples/lms_base_agent) — minimal chat with LM Studio
- [examples/lms_pro_agent](examples/lms_pro_agent) — more advanced local-model agent
- [examples/or_chat_agent](examples/or_chat_agent) — chat via OpenRouter
- [e2e_tests/e2e_format_text_agent](e2e_tests/e2e_format_text_agent) — JSON-schema structured output via OpenRouter

### Structured output (JSON schema)

Ask for JSON that conforms to a schema instead of free text. Two lines do
most of the work, both highlighted below:

- `(&jsonschema.Reflector{DoNotReference: true}).Reflect(&Person{})` —
  generates the JSON schema straight from a Go struct, so there is no
  hand-written schema to keep in sync.
- `jsonschema:"description=..."` on each field — the description lands in the
  generated schema and makes the contract readable to the model.

A JSON answer is far easier to put to work than free text: parse it, pass it
to another tool, or persist it without fragile string scraping. Validation is
simpler too — every field already carries a type, and the `description` can
spell out the expected value range, which the model generally respects.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/joho/godotenv"

	"rellm/pkg/rellm"
)

// The struct tags describe the schema.
type Person struct {
	Name string `json:"name" jsonschema:"description=Full name of the person"`
	Age  int    `json:"age"  jsonschema:"description=Age in years"`
	City string `json:"city" jsonschema:"description=City of residence"`
}

func main() {
	// Schema generated from the struct tags above.
	textFormat := rellm.TextFormat{
		Type:   "json_schema",
		Name:   "person",
		Strict: true,
		Schema: (&jsonschema.Reflector{DoNotReference: true}).Reflect(&Person{}),
	}

	_ = godotenv.Load() // reads OPENROUTER_API_KEY from .env
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		panic("OPENROUTER_API_KEY is not set")
	}

	provider, err := rellm.NewOpenRouterProvider(apiKey, "openai/gpt-5.6-luna")
	if err != nil {
		panic(err)
	}

	agent, err := rellm.NewAgentBuilder().
		WithProvider(provider).
		WithAgentName("StructuredOutputAgent").
		WithMaxAgentSteps(20).
		WithConversation(rellm.NewInMemoryConversation()).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		WithSystemMessage("You are an assistant that extracts structured information from user text and returns only valid JSON matching the requested schema.").
		WithTextFormat(textFormat).
		Build()
	if err != nil {
		panic(err)
	}

	prompt, err := rellm.NewPromptBuilder().
		WithMessage("I am John Snow from Winterfell, I have 100 years...").
		Build()
	if err != nil {
		panic(err)
	}

	finalReport, err := agent.Execute(context.Background(), prompt)
	if err != nil {
		panic(err)
	}

	var person Person
	err = json.Unmarshal([]byte(finalReport.Message), &person)
	if err != nil {
		panic(err)
	}
	fmt.Printf("after unmarshal to struct: %+v\n", person)
}
```

> **Local models (LM Studio):** a small model served by LM Studio may ignore
> the `text.format` field. If the output drifts, embed a stringified copy of
> the schema in the system prompt with an extra instruction to return raw
> JSON only (no markdown) — see how the LMS agent does it in
> [e2e_tests/e2e_format_text_agent/agent.go](e2e_tests/e2e_format_text_agent/agent.go).

### Adding tools to your agent

Tools define how the agent can interact with your system. Implement the
[`Toolset`](pkg/rellm/agent.go) interface and register it with
`WithToolset` on the builder:

```go
type Toolset interface {
    // Definitions returns the tool definitions advertised to the model.
    Definitions() []rellm.ToolDefinition

    // Dispatch executes a tool call requested by the model.
    Dispatch(ctx context.Context, name string, arguments json.RawMessage) (rellm.ToolCallResult, error)
}
```

**Recommendation: keep the functionality separate from the toolset implementation.**
Write your logic as a plain struct with methods (or plain functions), then wrap
it in a thin `Toolset` that only translates between the model and your code.
The official sample for a toolset is the Calculator, shipped in
[pkg/toolsets](pkg/toolsets): [calculator.go](pkg/toolsets/calculator.go) is
the functionality, [calculatortoolset.go](pkg/toolsets/calculatortoolset.go)
is the adapter.

```go
// calculator.go — your functionality: a plain struct, no rellm types involved.
type Calculator struct{}

func (c *Calculator) Add(a, b float64) float64 { return a + b }
func (c *Calculator) Sub(a, b float64) float64 { return a - b }
func (c *Calculator) Mul(a, b float64) float64 { return a * b }
```

```go
// calculatortoolset.go — a thin adapter that exposes Calculator to the agent.
type CalculatorToolset struct {
    impl *Calculator
}

func NewCalculatorToolset(c *Calculator) *CalculatorToolset {
    return &CalculatorToolset{impl: c}
}

var _ rellm.Toolset = (*CalculatorToolset)(nil)

func (t *CalculatorToolset) Definitions() []rellm.ToolDefinition {
    return []rellm.ToolDefinition{
        {
            Type:        "function",
            Name:        "Calculator_Add",
            Description: "Add two numbers.",
            Parameters: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "a": map[string]any{"type": "number"},
                    "b": map[string]any{"type": "number"},
                },
                "required": []string{"a", "b"},
            },
        },
        {
            Type:        "function",
            Name:        "Calculator_Sub",
            Description: "Subtract b from a.",
            Parameters: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "a": map[string]any{"type": "number"},
                    "b": map[string]any{"type": "number"},
                },
                "required": []string{"a", "b"},
            },
        },
        {
            Type:        "function",
            Name:        "Calculator_Mul",
            Description: "Multiply two numbers.",
            Parameters: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "a": map[string]any{"type": "number"},
                    "b": map[string]any{"type": "number"},
                },
                "required": []string{"a", "b"},
            },
        },
    }
}

func (t *CalculatorToolset) Dispatch(_ context.Context, name string, args json.RawMessage) (rellm.ToolCallResult, error) {
    var a struct {
        A, B float64
    }
    if err := json.Unmarshal(args, &a); err != nil {
        return rellm.ToolCallResult{Err: err}, nil
    }

    switch name {
    case "Calculator_Add":
        return rellm.ToolCallResult{Value: t.impl.Add(a.A, a.B)}, nil
    case "Calculator_Sub":
        return rellm.ToolCallResult{Value: t.impl.Sub(a.A, a.B)}, nil
    case "Calculator_Mul":
        return rellm.ToolCallResult{Value: t.impl.Mul(a.A, a.B)}, nil
    default:
        return rellm.ToolCallResult{}, fmt.Errorf("unknown tool: %s", name)
    }
}
```

This separation pays off in testing: `Calculator` is tested as ordinary Go code,
with no model, no JSON schema, and no agent in the loop. The toolset adapter is
the only part that needs a `Toolset`-level test. See
[examples/lms_math_agent](examples/lms_math_agent) for a complete agent built
around the Calculator toolset.

A `Toolset` implementation is simple, repetitive work — one definition and one
case per method. It works best when generated by a coding agent: give it your
struct and let it produce the adapter, then review and test the result.

For a general introduction to tool calling, see
[Tool & Function Calling (OpenRouter docs)](https://openrouter.ai/docs/guides/features/tool-calling).

## Tips

- **Explore from `rellm.New...`.** The API is self-explanatory: type
  `rellm.New` and your IDE lists the constructors — providers, conversations,
  prompts, agents — each returning the type you need next. The builders
  (`NewAgentBuilder`, `NewPromptBuilder`, ...) walk you through the options, so
  intellisense is the documentation. Ergonomic and easygoing — and for agentic
  coding it just works: a coding agent explores the API the same way your IDE
  does.
- **Building a custom provider?** Start from the built-ins — read
  [provider_openai.go](pkg/rellm/provider_openai.go) and
  [provider_openrouter.go](pkg/rellm/provider_openrouter.go). If your backend
  conforms to the Responses API, the shared `Std` helpers do most of the work:
  `StdSendResponsesAPI` (transport), `StdToConversationElements` (parsing) and
  `StdToProviderRepresentation` (serialization). What is left is your URL,
  headers, and a `ProviderTag` of your own.
- **The agent runs only while you wait.** `Ask` and `Execute` are fully
  synchronous — when they return, everything is done, nothing runs in the
  background. The stage before and after each call belongs to you: prepare the
  prompt, then post-process the report before asking again.
- **Timeouts:** pass a `context.Context` with a deadline to `Ask`/`Execute`, or
  inject an `http.Client` with a `Timeout` via
  `NewLMStudioProviderWithHTTPClient`, `NewOpenRouterProviderWithHTTPClient`
  or `NewOpenAIProviderWithHTTPClient`.

## External links

- [OpenAI Responses API specification](https://developers.openai.com/api/reference/overview)
- [OpenAI OpenAPI specification (GitHub)](https://github.com/openai/openai-openapi)
- [OpenRouter Responses API reference](https://openrouter.ai/docs/api/reference/responses/overview)
- [Tool & Function Calling (OpenRouter docs)](https://openrouter.ai/docs/guides/features/tool-calling)
- [LM Studio: Responses API support announcement](https://lmstudio.ai/blog/lmstudio-v0.3.29)
- [LM Studio CLI](https://lmstudio.ai/docs/cli)
