# AGENTS.md

This file provides guidance for agentic coding agents operating in this repository.

## Project Overview

Clock is a lightweight visual scheduling framework based on Go cron, supporting DAG task dependencies and bash command execution. Frontend and backend are packaged into a single binary via `go:embed`.

## Build Commands

### Backend (Go)

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/service/

# Run tests matching a pattern
go test -v -run "TestExecutor" ./internal/service/

# Format code
go fmt ./...

# Build for Linux (AMD64)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o clock cmd/clock/main.go

# Build using Makefile
make build    # Linux binary
make mac      # macOS binary
make win      # Windows binary
make all      # Build all platforms + frontend
```

### Frontend (Vue 3 + TypeScript)

```bash
cd server/webapp

# Install dependencies
npm install

# Development mode (hot reload)
npm run dev

# Production build (outputs to web/dist)
npm run build

# Type checking only
vue-tsc --noEmit
```

## Code Style Guidelines

### Go

**Imports**
- Standard library imports first, third-party imports second, local imports third
- Separate groups with blank lines
- Use `clock/` prefix for local imports (e.g., `"clock/internal/domain"`)

```go
import (
    "bufio"
    "bytes"
    "context"
    "errors"
    "fmt"

    "github.com/labstack/echo/v4"

    "clock/internal/domain"
    "clock/internal/logger"
)
```

**Naming Conventions**
- Structs: `PascalCase` (e.g., `TaskHandler`, `Executor`)
- Interfaces: `PascalCase` with `er` suffix where appropriate (e.g., `Repository`, `Service`)
- Variables: `camelCase` (e.g., `taskRepo`, `runID`)
- Constants: `PascalCase` or `camelCase` depending on scope (package-level constants use PascalCase)
- Acronyms: preserve original casing (e.g., `ID`, `URL`, `UUID` not `Id`, `Url`, `Uuid`)
- Private fields: `camelCase` starting with letter (e.g., `taskRepo`, `runningMu`)
- Package name: short, lowercase, no underscores

**Error Handling**
- Use `errors.New()` for static error messages
- Use `fmt.Errorf("context: %w", err)` for wrapping errors with context
- Return errors early; avoid nested error handling
- Log errors at the call site with `logger.Errorf("context: %v", err)`
- Use custom error types (`apperrors.AppError`) for domain-specific errors

```go
// Good - early return with error
if err != nil {
    logger.Errorf("[GetTask] failed: %v", err)
    return HandleError(c, err)
}

// Good - wrapped error with context
return fmt.Errorf("failed to get task: %w", err)
```

**Struct Tags**
- GORM tags: `gorm:"primaryKey"`, `gorm:"index:idx_cid"`, `gorm:"default:1"`
- JSON tags: `json:"fieldName"`

**Concurrency**
- Use `sync.RWMutex` for read-write locks
- Always use `defer` to unlock mutexes
- Use `sync.WaitGroup` for waiting on goroutines

```go
e.runningMu.RLock()
defer e.runningMu.RUnlock()
```

**Context**
- Pass `context.Context` as first parameter when needed
- Use `context.WithCancel`, `context.WithTimeout` for cancellation/timeout

### TypeScript/Vue

**TypeScript**
- Use strict mode (enabled in tsconfig.json)
- Prefer `interface` over `type` for object shapes
- Use explicit types; avoid `any`
- Use `Record<K, V>` for map-like objects

```typescript
// Good
interface Task {
  tid: number
  cid: number
  name: string
  status: number
}

// Good - Record for maps
const statusMap: Record<number, string> = { 1: 'pending', 2: 'running' }
```

**Vue Components**
- Use `<script setup lang="ts">` syntax
- Define props with `defineProps<{...}>()` or `withDefaults`
- Use `ref()` for primitives, `reactive()` for objects
- Prefer Composition API over Options API

```vue
<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { Task } from '@/types/model'

const loading = ref(false)
const list = ref<Task[]>([])
const form = reactive({
  name: '',
  command: ''
})
</script>
```

**Imports**
- Use path alias `@/` for `src/` (configured in tsconfig.json)
- Import types explicitly with `import type`

```typescript
import { getTasks, putTask } from '@/api/task'
import type { Task, Container } from '@/types/model'
```

**Styling**
- Use SCSS with CSS custom properties (variables)
- Use `scoped` attribute for component styles
- Use Element Plus components with PascalCase tags

## Project Structure

```
clock/
├── cmd/clock/
│   └── main.go           # Entry point
├── internal/
│   ├── config/           # TOML config loading
│   ├── domain/           # Core models (Task, Container, Relation, TaskLog, Message)
│   ├── errors/           # Custom error types
│   ├── handler/          # Echo HTTP handlers
│   ├── logger/           # Zap logger wrapper
│   ├── middleware/       # JWT auth, request logging
│   ├── repository/      # GORM data access layer
│   ├── router/           # Route registration
│   └── service/          # Business logic (SchedulerService, Executor, StreamHub)
├── pkg/util/             # Utility functions
├── server/webapp/        # Vue 3 frontend
│   └── src/
│       ├── api/          # Axios API clients
│       ├── stores/       # Pinia stores
│       ├── types/        # TypeScript type definitions
│       └── views/        # Vue page components
├── configs/              # Configuration files
└── Makefile              # Build automation
```

## Database

- Default: SQLite
- Also supports: MySQL, PostgreSQL
- Configured via `[storage]` section in `configs/config.toml`
- ORM: GORM with `glebarez/sqlite` driver

## Key Patterns

### SSE Real-time Updates
- `StreamHub` manages subscriber channels
- `Executor` publishes `StreamEvent` (task_start, task_end, stdout, stderr) to all subscribers
- Clients connect via `/v1/message` SSE endpoint

### DAG Execution
- `Executor.runStageTasks()` uses Kahn's algorithm for topological sorting
- Each stage executes all tasks with in-degree 0 in parallel
- After completion, remove executed nodes and their outgoing edges

### JWT Authentication
- Token passed via Header (`Authorization: Bearer <token>`), Cookie, or Query parameter
- Configured in `[auth]` section of `configs/config.toml`

## Testing

- Go tests use standard `testing` package
- Test files: `*_test.go` suffix
- No test files found in current codebase; add tests when modifying functionality

## IDE Recommendations

- VSCode with Go extension, Volar for Vue
- JetBrains GoLand / WebStorm
- Enable format on save for Go (`gofmt`) and TypeScript/Prettier
