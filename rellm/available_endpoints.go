package rellm

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type OpenRouterEndpoint struct {
}

func (e *OpenRouterEndpoint) GetUrl() string {
	return "https://openrouter.ai/api/v1/responses"
}

func (e *OpenRouterEndpoint) getApiKey() string {
	err := godotenv.Load()
	if err != nil {
		panic("cannot load .env: " + err.Error())
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		panic("missing apikey for OPENROUTER_API_KEY")
	}

	return apiKey
}

func (e *OpenRouterEndpoint) GetHttpHeader() http.Header {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Set("Authorization", "Bearer "+e.getApiKey())
	return header
}

type CustomResponseEndpoint struct {
	baseUrl             string
	port                string
	responseApiEndpoint string
	httpHeader          http.Header
}

func NewCustomResponseEndpoint(baseUrl, port, responseApiEndpoint string, httpHeader http.Header) *CustomResponseEndpoint {
	return &CustomResponseEndpoint{
		baseUrl:             baseUrl,
		port:                port,
		responseApiEndpoint: responseApiEndpoint,
		httpHeader:          httpHeader,
	}
}

func (e *CustomResponseEndpoint) GetUrl() string {
	return e.baseUrl + ":" + e.port + e.responseApiEndpoint
}

func (e *CustomResponseEndpoint) GetHttpHeader() http.Header {
	if e.httpHeader != nil {
		return e.httpHeader
	}

	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	return header
}
