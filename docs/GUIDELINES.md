# Go Development Best Practices (Agentic Coding Guide)

This document defines practical rules and conventions for writing maintainable, scalable Go (Golang) applications. It is designed to guide both humans and AI agents.

---

# 1. Architectural Separation

## 1.1 Core Layers

Structure your application into three main layers:

### 1. Infrastructure Layer (Execution Environment)
Responsible for **how the application is hosted and executed**:
- Microservice wiring (service startup, dependency injection)
- Lambda / serverless handlers
- CLI / OS binaries
- Process lifecycle
- Configuration loading (env, flags)

Characteristics:
- Entry point of the application
- Wires dependencies together
- NO business logic

❗ Important:
Infrastructure here does NOT mean database or external systems implementation.

---

### 1.2 Transport Layer
Responsible for communication with the outside world:
- HTTP handlers (REST)
- gRPC / Protobuf
- Message queue consumers/producers

Responsibilities:
- Request parsing
- Input validation
- Response formatting
- Calling business logic

Rules:
- MUST NOT contain business logic
- SHOULD be thin and declarative

---

### 1.3 Business Logic Layer (Domain)
Core of the application:
- Business rules
- Decision making
- State transitions

Characteristics:
- Independent of infrastructure and transport
- Testable in isolation
- Pure logic whenever possible

---

## 1.2 External Systems (Separate Packages)

External systems such as:
- Databases
- External APIs
- Message brokers

MUST live in their own packages (e.g. `internal/postgres`, `internal/redis`, `internal/httpclient`).

Rules:
- They implement interfaces defined in the business layer
- They are injected into services
- They are NOT part of the infrastructure layer

---

## 1.3 Dependency Direction

Dependencies must flow inward:

```
Transport -> Business Logic <- External Implementations
             ^
             |
       Interfaces
```

- Business logic defines interfaces
- External packages implement them
- Infrastructure layer wires everything together

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

- External dependencies MUST be hidden behind interfaces
- Interfaces SHOULD be small and focused
- Avoid "god interfaces"

---

## 4.2 Interface Placement (IMPORTANT FOR GO)

In Go, interfaces should be defined **where they are used**, not where they are implemented.

### ✅ Correct

```go
// business/service.go

type UserRepository interface {
    GetByID(id string) (User, error)
}

type Service struct {
    repo UserRepository
}
```

### ❌ Incorrect

```go
// infrastructure/repo.go

type UserRepository interface {
    GetByID(id string) (User, error)
}
```

---

## 4.3 When to Use Interfaces

Use interfaces ONLY when:

- There are multiple implementations
- You need to mock for testing
- You are decoupling business logic from external systems

Do NOT use interfaces:
- For single implementations without a clear need
- Prematurely

---

# 5. External Interaction Rule

Any component that interacts with the external world MUST:

- Live in its own package
- Implement an interface defined in business logic
- Be injected into services

Example:

```go
// business layer

type PaymentGateway interface {
    Charge(amount int) error
}
```

```go
// external package

type StripeGateway struct {}

func (s *StripeGateway) Charge(amount int) error {
    return nil
}
```

---

# 6. Transport Layer Rules

## 6.1 HTTP

Handlers should:
- Parse request
- Validate input
- Call service
- Return response

Example:

```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    req := parseRequest(r)

    err := h.service.CreateUser(req)
    if err != nil {
        writeError(w, err)
        return
    }

    writeResponse(w)
}
```

---

## 6.2 Message Queues

Consumers should:
- Deserialize message
- Validate
- Call business logic

Producers should:
- Serialize domain events
- Publish via interface

---

# 7. Business Logic Rules

- No HTTP, DB, or framework dependencies
- Accept interfaces, not implementations
- Focus on domain behavior

Example:

```go
func (s *Service) CreateUser(req CreateUserRequest) error {
    if req.Email == "" {
        return errors.New("email required")
    }

    return s.repo.Save(User{Email: req.Email})
}
```

---

# 8. External Packages (DB, APIs, Queues)

- Live outside business logic (e.g. `internal/postgres`)
- Implement interfaces from business layer
- Contain all technical details (SQL, HTTP calls, etc.)
- MUST NOT contain business decisions

Example:

```go
type PostgresUserRepository struct {}

func (r *PostgresUserRepository) Save(user User) error {
    // SQL implementation
    return nil
}
```

---

# 9. Infrastructure Layer (Wiring Example)

This layer connects everything together (composition root).

Example:

```go
func main() {
    fsAccessPoint := NewSingleDirectoryFs("/tmp")
    ticketApi := NewJiraApi()

    agent := NewAgent(fsAccessPoint, ticketApi)
    server := NewServer(agent)

    startHTTPServer(server)
}
```

Responsibilities:
- Create concrete implementations
- Inject dependencies into business logic
- Start application (HTTP, worker, lambda, etc.)

---

## 9.1 Example Transport + Wiring Flow

```go
type Server struct {
    agent *Agent
}

func NewServer(agent *Agent) *Server {
    return &Server{agent: agent}
}

func (s *Server) httpHandler(resp http.ResponseWriter, req *http.Request) {
    // 1. Read request
    var msg Message
    err := decodeRequest(req, &msg)
    if err != nil {
        writeHTTPError(resp, err)
        return
    }

    // 2. Call business logic
    data, err := s.agent.ProcessMessage(msg)
    if err != nil {
        writeHTTPError(resp, err)
        return
    }

    // 3. Write response
    err = writeResponse(resp, data)
    if err != nil {
        writeHTTPError(resp, err)
        return
    }
}
```

### Key Rules Demonstrated

- Transport only:
    - parses input
    - calls business logic
    - formats output
- No business decisions in handler
- Errors from business logic are translated to HTTP errors
- Agent (business logic) is unaware of HTTP

---

# 10. General Coding Rules

## 10.1 Readability First

- Prefer clarity over cleverness
- Use meaningful names
- Avoid abbreviations

## 10.2 Error Handling

- Always handle errors explicitly
- Avoid panic in business logic

## 10.3 Early Returns

Reduce nesting:

```go
if err != nil {
    return err
}
```

---

# 11. Summary Checklist

Before writing code, ensure:

- [ ] Clear separation of layers
- [ ] Functions <= 20 lines (ideally <= 10)
- [ ] Complexity <= 10
- [ ] Interfaces used correctly
- [ ] No business logic in transport
- [ ] External systems in separate packages
- [ ] External systems hidden behind interfaces
- [ ] Infrastructure only wires dependencies
- [ ] Dependencies flow inward

---

# 12. Guiding Principle

> Keep code small, simple, decoupled, and replaceable.

This ensures the system remains maintainable, testable, and scalable over time.

