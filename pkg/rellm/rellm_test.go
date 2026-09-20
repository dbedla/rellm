package rellm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"github.com/dbedla/rellm/internal/examplesutils"
	"github.com/dbedla/rellm/pkg/rellm"
	"testing"

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

// expectedStepStats unmarshals the usage blocks of the given golden responses
// into the StepStats a report should carry, one per successful provider call.
func expectedStepStats(t *testing.T, responses ...string) []rellm.StepStat {
	t.Helper()
	stats := make([]rellm.StepStat, len(responses))
	for i, resp := range responses {
		var r rellm.ResponsesAPIResp
		if err := json.Unmarshal([]byte(resp), &r); err != nil {
			t.Fatalf("failed to unmarshal golden response: %v", err)
		}
		stats[i] = rellm.StepStat{APIUsage: r.Usage}
	}
	return stats
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
	req.Temperature = new(float64(0.5))

	examplesutils.InspectWithReqLog(req)
}

func SetParametersForTest(req *rellm.ResponsesAPIReq) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: rellm.ReasoningEffortMedium}
	req.Temperature = new(float64(0.5))
}
