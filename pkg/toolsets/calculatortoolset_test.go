package toolsets

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestCalculatorToolsetDispatch(t *testing.T) {
	ts := NewCalculatorToolset(&Calculator{})
	ctx := context.Background()

	cases := []struct {
		name    string
		tool    string
		args    string
		want    any
		wantErr bool
	}{
		{"Add", "Calculator_Add", `{"a":2,"b":3}`, float64(5), false},
		{"Sub", "Calculator_Sub", `{"a":2,"b":3}`, float64(-1), false},
		{"Mul", "Calculator_Mul", `{"a":2,"b":3}`, float64(6), false},
		{"Unknown tool", "Calculator_Div", `{"a":2,"b":3}`, nil, true},
		{"Bad JSON", "Calculator_Add", `{oops`, nil, false}, // unmarshal error goes into result.Err
	}
	for _, tc := range cases {
		res, err := ts.Dispatch(ctx, tc.tool, json.RawMessage(tc.args))
		if tc.wantErr {
			if err == nil {
				t.Errorf("%s: expected error, got none", tc.name)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}
		if tc.name == "Bad JSON" {
			if res.Err == nil {
				t.Errorf("%s: expected res.Err, got none", tc.name)
			}
			continue
		}
		if !reflect.DeepEqual(res.Value, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, res.Value, tc.want)
		}
	}
}
