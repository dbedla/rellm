package rellm

import (
	"net/http"
)

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
	port := ""
	if e.port != "" {
		port = ":" + e.port
	}

	return e.baseUrl + port + e.responseApiEndpoint
}

func (e *CustomResponseEndpoint) GetHttpHeader() http.Header {
	if e.httpHeader != nil {
		return e.httpHeader
	}

	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	return header
}
