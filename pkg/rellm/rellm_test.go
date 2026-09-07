package rellm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"

	"github.com/stretchr/testify/mock"
)

const (
	testBaseUrl              = "http://127.0.0.1"
	testPort                 = "1234"
	testReasoningEffort      = rellm.ReasoningEffortLow
	testTemperature          = 0.5
	testDefaultMaxAgentSteps = rellm.DefaultMaxAgentSteps
)

func baseRequestMatch(req *http.Request) bool {
	return req.Method == http.MethodPost && req.ContentLength != 0
}

type HTTPDoMock struct {
	mock.Mock
}

func (h *HTTPDoMock) Do(req *http.Request) (*http.Response, error) {
	args := h.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

type ToolsetMock struct {
	mock.Mock
}

var _ rellm.Toolset = (*ToolsetMock)(nil)

func (m *ToolsetMock) Definitions() []rellm.ToolDefinition {
	args := m.Called()
	return args.Get(0).([]rellm.ToolDefinition)
}

func (m *ToolsetMock) Dispatch(ctx context.Context, name string, arguments json.RawMessage) (rellm.ToolCallResult, error) {
	args := m.Called(ctx, name, arguments)
	return args.Get(0).(rellm.ToolCallResult), args.Error(1)
}

func SetParametersWithReqLog(req *rellm.ResponsesAPIReq) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: rellm.ReasoningEffortMedium}
	req.Temperature = new(float32(0.5))

	examplesutils.InspectWithReqLog(req)
}

func SetParametersForTest(req *rellm.ResponsesAPIReq) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: rellm.ReasoningEffortMedium}
	req.Temperature = new(float32(0.5))
}
