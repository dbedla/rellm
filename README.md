# What is rellm?
rellm is an experimental go library for building custom agents.

## What endpoint is used
Right now rellm use response api proposed by OpenAI. But do not worry, this endpoint is implemented by other providers. LM Studio provide it also, so it can be used with models run locally. OpenRouter is a unified API gateway and also provide this endpoint so it.

## Can it work with local models?
Yes.\
Library allows work with both **local** and **external** models.

The `./cmd/basic_agent` example demonstrates how to use the `google/gemma-4-12b-qat` model running on a local machine. (Models is served by LM Studio)

**Configuration:**

- **Model:** `google/gemma-4-12b-qat`
- **Host:** `http://127.0.0.1`
- **Port:** `1234`
- **API Endpoint:** `/v1/responses`


## Can it work with any model?
It can work with all model providers that implement responses api endpoint. To check which model works best with your agent some benchmarks and experiments are needed.\
Recommendations:
 - local models -> LM Studio
 - top high-end paid models -> OpenRouter

## Quick start guide (run example in cmd/basic_agent)
[Install LM Studio](https://lmstudio.ai/download) on your machine it will be used to run local models.
Then go to the root of this repository and execute commands
```bash
make lms-set-gemma-4-12b
make go-build
./output/bin/base_agent
```
# External links
 - [Open Router responses api](https://openrouter.ai/docs/api/reference/responses/overview)
 - [LM Studio: Use OpenAI's Responses API with local models](https://lmstudio.ai/blog/lmstudio-v0.3.29)
 - [LM Studio cli](https://lmstudio.ai/docs/cli)