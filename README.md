# What is `rellm`?
`rellm` is an Go lightweight framework for building custom agents based on the Responses API endpoint.

## What endpoint is used?
Right now, `rellm` uses the Responses API proposed by OpenAI. But do not worry, this endpoint is implemented by other providers too. LM Studio provides it, so it can be used with models run locally. OpenRouter is a unified API gateway and also implements this endpoint.
Local model recomendations:
- ornith-1.5-35b-a3b and previous version for slower machine deepreinforce-ai/ornith-1.0-35b
- google/gemma-4-26b-a4b
- gemma-4-26b-a4b
- qwen/qwen3.8-27b


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
It can work with all providers that implement the Responses API endpoint. To check which model works best with your agent, some benchmarks and experiments are needed.

Local models -> LM Studio 
- ornith-1.5-35b-a3b 
- deepreinforce-ai/ornith-1.0-35b
- google/gemma-4-26b-a4b
- qwen/qwen3.8-27b
- meta/muse-glimmer
 
Benchmarks:
- https://openrouter.ai/benchmarks


## Quick start guide: build own agent
Reference usage of a library can be found:
- [base agent - chat functionaliuty](examples/lms_base_agent)
- [agent with toolset - limited support of file system operation in a restricted path](e2e_tests/e2e_fs_agent)

### How to extend an agent with user-defined functions?
If you want build agent and extend his interaction abilities you can use [Tool & Function Calling] (https://openrouter.ai/docs/guides/features/tool-calling) from `v1/responses` endpoint.
See [the fs_agent example](e2e_tests/e2e_fs_agent) for more details.

## Quick start guide (run example in `./examples/base_agent`)
[Install LM Studio](https://lmstudio.ai/download) on your machine; it will be used to run local models.
Then go to the root of this repository and execute the following commands:
```bash
# if you want use local model (gemma-4-26b-a4b)
make e2e-lms-env 

# builds binary :D 
make go-build

# just run it to see errors :D
# for openrouter as provider .env file is required
# OPENROUTER_API_KEY=sk-1234567890
# how to generate openrouter api key (available only for logged user): https://openrouter.ai/workspaces/default/keys
./output/bin/lms_endpoint_agent 
```
## External links
 - [Open Router responses api](https://openrouter.ai/docs/api/reference/responses/overview)
 - [LM Studio: Use OpenAI's Responses API with local models](https://lmstudio.ai/blog/lmstudio-v0.3.29)
 - [LM Studio cli](https://lmstudio.ai/docs/cli)
 - [OpenAI API · OpenAPI specification - github](https://github.com/openai/openai-openapi)
 - [OpenAI API Overview](https://developers.openai.com/api/reference/overview)