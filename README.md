# What is `rellm`?
`rellm` is an experimental Go library for building custom agents based on the Responses API endpoint.

## What endpoint is used?
Right now, `rellm` uses the Responses API proposed by OpenAI. But do not worry, this endpoint is implemented by other providers too. LM Studio provides it, so it can be used with models run locally. OpenRouter is a unified API gateway and also implements this endpoint.

## Can it work with local models?
Yes.
The library works with both **local** and **external** models.

The `./examples/base_agent` example demonstrates how to use the `google/gemma-4-12b-qat` model running on a local machine. (Models can be served by LM Studio)

**Configuration:**

- **Model:** `google/gemma-4-12b-qat`
- **Host:** `http://127.0.0.1`
- **Port:** `1234`
- **API Endpoint:** `/v1/responses`


## Can it work with any model?
It can work with all model providers that implement the Responses API endpoint. To check which model works best with your agent, some benchmarks and experiments are needed.

Recommendations:
 - Local models -> LM Studio
 - Top high-end paid models -> OpenRouter

## Quick start guide: build own agent
Reference usage of a library can be found:
- [base agent - chat functionaliuty](./examples/base_agent)
- [agent with toolset - limited support of file system operation in a restricted path](./examples/fs_agent)

### How to extend an agent with user-defined functions?
If you want build agent and extend his interaction abilities you can use [Tool & Function Calling] (https://openrouter.ai/docs/guides/features/tool-calling) from `v1/responses` endpoint.
See [the fs_agent example](./examples/fs_agent) for more details.

## Quick start guide (run example in `./examples/base_agent`)
[Install LM Studio](https://lmstudio.ai/download) on your machine; it will be used to run local models.
Then go to the root of this repository and execute the following commands:
```bash
make lms-set-gemma-4-12b
make go-build
./output/bin/base_agent
```
## External links
 - [Open Router responses api](https://openrouter.ai/docs/api/reference/responses/overview)
 - [LM Studio: Use OpenAI's Responses API with local models](https://lmstudio.ai/blog/lmstudio-v0.3.29)
 - [LM Studio cli](https://lmstudio.ai/docs/cli)