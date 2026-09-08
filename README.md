# rellm

`rellm` (from **Res**ponses API + **LLM** + endpoint) is a lightweight Go
framework for building custom agents on top of the OpenAI **Responses API**
(`/v1/responses`). Multiple providers implement this endpoint (LM Studio,
OpenRouter), which lets you freely switch between local and hosted models
to balance cost and result quality.

A key feature is easy custom toolset injection: you decide exactly how the
agent can interact with your system. More control over the toolset means
fewer unexpected side effects when the model hallucinates.

## Table of contents

- [Why small, specialized agents?](#why-small-specialized-agents)
- [Supported providers & models](#supported-providers--models)
- [Installation](#installation)
- [Quick start](#quick-start)
  - [Run the example agent](#run-the-example-agent)
  - [Build your own agent](#build-your-own-agent)
  - [Adding tools to your agent](#adding-tools-to-your-agent)
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

## Installation
Intentionaly left blank.

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

### Build your own agent

```go
package main

import (
    "context"

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
        WithConversationStorage(rellm.NewInMemoryStorage()).
        WithSystemMessage("You are a helpful assistant.").
        Build()
    if err != nil {
        panic(err)
    }

    response, err := agent.Ask(context.Background(), "Hello!")
    _ = response
}
```

See the full runnable version in [examples/endpoint_agent](examples/endpoint_agent),
which shows how to switch between providers with flags.
Other examples:

- [examples/lms_base_agent](examples/lms_base_agent) — minimal chat with LM Studio
- [examples/lms_pro_agent](examples/lms_pro_agent) — more advanced local-model agent
- [examples/or_chat_agent](examples/or_chat_agent) — chat via OpenRouter

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
The complete working version is
[examples/lms_math_agent](examples/lms_math_agent): `calculator.go` is the
functionality, `calculatortoolset.go` is the adapter — shown here trimmed:

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
    *Calculator
}

func NewCalculatorToolset(c *Calculator) *CalculatorToolset {
    return &CalculatorToolset{Calculator: c}
}

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
        // Definitions for Calculator_Sub and Calculator_Mul follow
        // the same pattern.
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
        return rellm.ToolCallResult{Value: t.Add(a.A, a.B)}, nil
    case "Calculator_Sub":
        return rellm.ToolCallResult{Value: t.Sub(a.A, a.B)}, nil
    case "Calculator_Mul":
        return rellm.ToolCallResult{Value: t.Mul(a.A, a.B)}, nil
    default:
        return rellm.ToolCallResult{}, fmt.Errorf("unknown tool: %s", name)
    }
}
```

This separation pays off in testing: `Calculator` is tested as ordinary Go code,
with no model, no JSON schema, and no agent in the loop. The toolset adapter is
the only part that needs a `Toolset`-level test.

A `Toolset` implementation is simple, repetitive work — one definition and one
case per method. It works best when generated by a coding agent: give it your
struct and let it produce the adapter, then review and test the result.

The framework itself follows this pattern — see
[`LimitedFileSystem`](pkg/agentsutils/limited_file_system.go) (functionality)
wrapped by [`FSToolset`](pkg/agentsutils/limited_file_system_toolset.go), used
by [the fs_agent example](e2e_tests/e2e_fs_agent): an agent restricted to
file-system operations within a safe path.

For a general introduction to tool calling, see
[Tool & Function Calling (OpenRouter docs)](https://openrouter.ai/docs/guides/features/tool-calling).

## External links

- [OpenAI Responses API specification](https://developers.openai.com/api/reference/overview)
- [OpenAI OpenAPI specification (GitHub)](https://github.com/openai/openai-openapi)
- [OpenRouter Responses API reference](https://openrouter.ai/docs/api/reference/responses/overview)
- [Tool & Function Calling (OpenRouter docs)](https://openrouter.ai/docs/guides/features/tool-calling)
- [LM Studio: Responses API support announcement](https://lmstudio.ai/blog/lmstudio-v0.3.29)
- [LM Studio CLI](https://lmstudio.ai/docs/cli)
