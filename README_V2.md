# rellm

`rellm` (from **Res**ponses API + **LLM** + endpoint) is a lightweight Go
framework for building custom agents on top of the OpenAI **Responses API**
(`/v1/responses`). Multiple providers implement this endpoint (LM Studio,
OpenRouter), which lets you freely switch between local and hosted models
to balance cost and result quality.

A key feature is easy custom toolset injection: you decide exactly how the
agent can interact with your system. More control over the toolset means
fewer unexpected side effects when the model hallucinates.

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
- [LMArena](https://lmarena.ai)

## Installation
Intentionaly left blank.

## Quick start

### Run the example agent

### Build your own agent

### Adding tools to your agent

## External links
