package toolsets

import "testing"

func TestCalculator(t *testing.T) {
	c := &Calculator{}
	cases := []struct {
		name string
		got  float64
		want float64
	}{
		{"Add", c.Add(2, 3), 5},
		{"Add negatives", c.Add(-2, 3), 1},
		{"Sub", c.Sub(2, 3), -1},
		{"Sub to zero", c.Sub(3, 3), 0},
		{"Mul", c.Mul(2, 3), 6},
		{"Mul by zero", c.Mul(2, 0), 0},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, tc.got, tc.want)
		}
	}
}
