# Changelog

## Unreleased

- Added image inputs for vision prompts: `PromptBuilder.WithImageURL`,
  `WithImageDataURL`, `WithImageBytes`, `WithImageFile`, and `WithImageFileID`,
  each taking an `ImageOption` carrying the doc-mentioned optional fields:
  `Detail` (defaults to `auto`), `FileID` (uploaded-file reference), and
  `CacheBreakpoint` (`PromptCacheBreakpoint`, an OpenAI-style prompt-cache
  boundary converted by OpenRouter for Anthropic-style providers); a message
  or at least one image is now required to build a prompt.
- Fixed `MessagePart.ImageURL` wire type: `input_image.image_url` is a string
  (URL or data URL), not an object — provider echoes of image parts previously
  parsed to empty content. `MessagePart` now also carries `file_id` and
  `prompt_cache_breakpoint`. **Breaking** for users of the removed `ImageURL`
  struct.

## v0.1.0 - 2026-09-20

- Added synchronous agents for the OpenAI Responses API.
- Added OpenAI, OpenRouter, and LM Studio providers, plus helpers for custom providers.
- Added conversations, tool calling, structured output, reasoning, and image handling.
- Added builders, examples, end-to-end tests, and a calculator toolset.
