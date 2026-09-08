package main

// Calculator is the functionality: a plain struct with methods,
// testable without any rellm types, model, or agent.
type Calculator struct{}

func (c *Calculator) Add(a, b float64) float64 { return a + b }

func (c *Calculator) Sub(a, b float64) float64 { return a - b }

func (c *Calculator) Mul(a, b float64) float64 { return a * b }
