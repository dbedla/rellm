package main

import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/dbedla/rellm/internal/examplesutils"
	"github.com/dbedla/rellm/pkg/rellm"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

const structuredOutputSysPrompt = `You are an assistant that extracts structured information from user text and returns only valid JSON matching the requested schema.`

func buildLMSStructuredOutputAgent() (*rellm.Agent, error) {
	p, err := rellm.NewLMStudioProvider("google/gemma-4-26b-a4b", "http://127.0.0.1", "1234")
	if err != nil {
		return nil, err
	}
	schema, err := json.Marshal(textFormat.Schema)
	if err != nil {
		return nil, err
	}
	lmsSysPrompt := structuredOutputSysPrompt + string(schema) + "\nONLY parsable json string allowed as result, no additional markdown formatting\n"
	return buildStructuredOutputAgent(p, "lms-agent", lmsSysPrompt)
}

func buildOpenAIStructuredOutputAgent() (*rellm.Agent, error) {
	p, err := newOpenAIProvider("gpt-5.6-luna")
	if err != nil {
		return nil, err
	}
	return buildStructuredOutputAgent(p, "openai-agent", structuredOutputSysPrompt)
}

func buildORGlm3flashStructuredOutputAgent() (*rellm.Agent, error) {
	p, err := newOpenRouterProvider("z-ai/glm-5.3-flash")
	if err != nil {
		return nil, err
	}
	return buildStructuredOutputAgent(p, "or-glm-5.3-flash-agent", structuredOutputSysPrompt)
}

func buildORLunaStructuredOutputAgent() (*rellm.Agent, error) {
	p, err := newOpenRouterProvider("openai/gpt-5.6-luna")
	if err != nil {
		return nil, err
	}
	return buildStructuredOutputAgent(p, "or-luna-agent", structuredOutputSysPrompt)
}

func buildStructuredOutputAgent(provider rellm.Provider, name string, sysPrompt string) (*rellm.Agent, error) {

	return rellm.NewAgentBuilder().
		WithAgentName(name).
		WithProvider(provider).
		WithMaxAgentSteps(20).
		WithConversation(rellm.NewInMemoryConversation()).
		WithSystemMessage(sysPrompt).
		WithTextFormat(textFormat).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()
}

func buildAgentForFlag(fl flag) (*rellm.Agent, error) {
	switch fl {
	case flag_LMS:
		return buildLMSStructuredOutputAgent()
	case flag_OpenRouterGlm53flash:
		return buildORGlm3flashStructuredOutputAgent()
	case flag_OpenRouterOpenAILuna:
		return buildORLunaStructuredOutputAgent()
	case flag_OpenAI:
		return buildOpenAIStructuredOutputAgent()
	default:
		return nil, fmt.Errorf("unknown flag provided: %s", fl)
	}
}

type flag string

const (
	flag_LMS                  flag = "--lms"
	flag_OpenRouterGlm53flash flag = "--or-glm53flash"
	flag_OpenRouterOpenAILuna flag = "--or-openai-luna"
	flag_OpenAI               flag = "--openai"
	flag_Invalid              flag = "NO_FLAG"
)

func help() {
	color.Yellow("allowed args:")
	color.Yellow("\t %s", flag_LMS)
	color.Yellow("\t %s", flag_OpenRouterGlm53flash)
	color.Yellow("\t %s", flag_OpenRouterOpenAILuna)
	color.Yellow("\t %s", flag_OpenAI)
}

func argsToFlag(args []string) flag {
	if len(args) != 2 {
		return flag_Invalid
	}

	f := flag(args[1])
	if f == flag_LMS || f == flag_OpenRouterGlm53flash || f == flag_OpenRouterOpenAILuna || f == flag_OpenAI {
		return f
	}

	return flag_Invalid
}

func newOpenRouterProvider(model rellm.Model) (rellm.Provider, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load .env: %w", err)
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing apikey for OPENROUTER_API_KEY")
	}

	return rellm.NewOpenRouterProvider(apiKey, model)
}

func newOpenAIProvider(model rellm.Model) (rellm.Provider, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load .env: %w", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing apikey for OPENAI_API_KEY")
	}

	return rellm.NewOpenAIProvider(apiKey, model)
}
