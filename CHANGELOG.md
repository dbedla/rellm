# Changelog

## Unreleased

- Improved marshaling performance: conversation elements' `MarshalJSON` and
  the wire marshal helpers now serialize typed payload structs in a single
  pass instead of round-tripping through `map[string]` and re-injecting the
  kind discriminator (`marshalWithKind` is gone). Persisted JSONL lines
  changed key order from alphabetical to struct declaration order
  (`kind`, `role`, `content`, ...) — parsing is unaffected, but
  byte-comparisons of stored history will differ.
- Added image inputs for vision prompts: `PromptBuilder.WithImageURL`,
  `WithImageDataURL`, `WithImageBytes`, `WithImageFile`, and `WithImageFileID`,
  each taking an `ImageOption` carrying the doc-mentioned optional fields:
  `Detail` (defaults to `auto`) and `CacheBreakpoint` (`PromptCacheBreakpoint`,
  an OpenAI-style prompt-cache boundary converted by OpenRouter for
  Anthropic-style providers); a file ID is set only through
  `WithImageFileID`'s argument because a wire part carries either
  `image_url` or `file_id`, never both; a message
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
