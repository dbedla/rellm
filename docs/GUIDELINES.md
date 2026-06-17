# Go Development Best Practices (Agentic Coding Guide)

This document defines practical rules and conventions for writing maintainable, scalable Go (Golang) applications. It is designed to guide both humans and AI agents.

---

# 1. Architectural Patterns

## 1.1 Dependency Injection
- Use dependency injection to keep components decoupled and testable.
- Inject dependencies via constructors or builders.

## 1.2 Builder Pattern
- For complex objects like `Agent` or `Endpoint`, use the Builder pattern to provide a clean and flexible initialization API.
- Example: `NewAgentBuilder().WithEndpoint(e).Build()`.

## 1.3 Repository/Storage Pattern
- Abstract data persistence (e.g., `ConversationStorage`) behind interfaces or dedicated structs to decouple business logic from storage details.

---

# 2. Function Size Guidelines

Function length is a strong indicator of readability and maintainability.

| Lines | Status      | Action               |
|-------|-------------|----------------------|
| 0–10  | ✅ Good      | Ideal size           |
| 11–15 | ⚠️ Warning  | Consider refactoring |
| 16–20 | ⚠️ Critical | Refactor soon        |
| >20   | ❌ Bad       | MUST refactor        |

## 2.1 Refactoring Strategies

- Extract smaller functions
- Remove nested logic
- Use early returns
- Separate responsibilities

---

# 3. Function Complexity Guidelines

Use cyclomatic complexity as a guideline.

| Complexity | Status        | Action               |
|------------|---------------|----------------------|
| 0–5        | ✅ Good        | Ideal                |
| 6–10       | ⚠️ Acceptable | Avoid further growth |
| >10        | ❌ Bad         | MUST refactor        |

## 3.1 Reducing Complexity

- Split logic into smaller functions
- Replace condition chains with polymorphism
- Use strategy pattern
- Avoid deep nesting

---

# 4. Interfaces and Abstractions

## 4.1 General Rules

- External dependencies MUST be hidden behind interfaces.
- Interfaces SHOULD be small and focused (e.g., `Toolset`, `ClientHttpDo`).
- Avoid "god interfaces".

## 4.2 Interface Placement (IMPORTANT FOR GO)

In Go, interfaces should be defined **where they are used**, not where they are implemented.

### ✅ Correct

```go
// pkg/rellm/agent.go

type Toolset interface {
    BuildTools() []Tool
    DispatchTools(name string, callID string, arguments json.RawMessage) (FunctionCallResp, bool)
}
```

## 4.3 When to Use Interfaces

Use interfaces ONLY when:
- There are multiple implementations (e.g., different LLM endpoints).
- You need to mock for testing.
- You are decoupling business logic from external systems.

Do NOT use interfaces:
- For single implementations without a clear need.
- Prematurely.

---

# 5. External Interaction Rule

Any component that interacts with the external world (HTTP clients, file system, LLM APIs) MUST:
- Live in its own package or be properly abstracted.
- Implement an interface defined where it is used.
- Be injected into services.

---

# 6. General Coding Rules

## 6.1 Readability First
- Prefer clarity over cleverness.
- Use meaningful names.
- Avoid abbreviations unless they are standard (e.g., `ID`, `URL`, `LLM`).

## 6.2 Error Handling
- Always handle errors explicitly.
- Use `errors.Join` or `%w` for error wrapping.
- Avoid `panic`.

## 6.3 Early Returns
Reduce nesting by returning early:

```go
if err != nil {
    return err
}
```

## 6.4 Logging
- Use structured logging (e.g., `zerolog`).
- Log at appropriate levels (`Info`, `Debug`, `Error`, `Warn`).
- Include relevant context (e.g., `component`, `agentName`).

---

# 7. Tooling and Automation

Always use `Makefile` for common tasks:
- `make go-build`: Build the project.
- `make go-test`: Run all tests.
- `make go-lint`: Run linters (`go vet`, `golangci-lint`).
- `make go-cyclo`: Check cyclomatic complexity.

---

# 8. Summary Checklist

Before writing code, ensure:
- [ ] Functions <= 20 lines (ideally <= 10).
- [ ] Complexity <= 10.
- [ ] Interfaces used correctly (defined at usage site).
- [ ] External systems hidden behind interfaces.
- [ ] Dependencies flow inward.
- [ ] Use `make` for building and testing.

---

# 9. Guiding Principle

> Keep code small, simple, decoupled, and replaceable.

This ensures the system remains maintainable, testable, and scalable over time.
