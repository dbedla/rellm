package rellm_test

import (
	"net/http"
	"rellm/pkg/agentsutils"
	"rellm/pkg/rellm"

	"github.com/stretchr/testify/mock"
)

const (
	testBaseUrl                                      = "http://127.0.0.1"
	testPort                                         = "1234"
	testResponsesApiEndpoint                         = "/v1/responses"
	testReasoningEffort                              = "low"
	testTemperature                                  = 0.5
	TestDefaultMaxToolsIterationWithoutReturnMessage = 5
)

func baseRequestMatch(req *http.Request) bool {
	return req.URL.String() == testBaseUrl+":"+testPort+testResponsesApiEndpoint &&
		req.Method == "POST"
}

func setTestReasoningAndTemperature(req *rellm.ResponsesApiReq) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: testReasoningEffort}
	req.Temperature = testTemperature
}

type HttpDoMock struct {
	mock.Mock
}

func (h *HttpDoMock) Do(req *http.Request) (*http.Response, error) {
	args := h.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func SetParametersWithReqLog(req *rellm.ResponsesApiReq) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: "medium"}
	req.Temperature = 0.5

	agentsutils.InspectWithReqLog(req)
}

func SetParametersForTest(req *rellm.ResponsesApiReq) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: "medium"}
	req.Temperature = 0.5
}
