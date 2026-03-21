# 测试体系设计文档

**日期：** 2026-03-21
**主题：** Clock 项目测试体系建设

---

## 1. 目标

为 Clock 项目建立可持续的测试体系，覆盖核心模块，确保代码质量和重构安全性。

---

## 2. 测试框架

- **测试框架：** Go 标准库 `testing`
- **断言库：** `github.com/stretchr/testify/assert`
- **Mock 方式：** 接口 + manual mock（不引入 mock 框架，保持简单）

**依赖安装：**
```bash
go get github.com/stretchr/testify
```

---

## 3. 测试文件结构

采用**分散式**结构，测试文件与被测文件同目录：

```
internal/service/
├── executor.go
├── executor_test.go      # Executor 单元测试
├── stream_hub.go
├── stream_hub_test.go    # StreamHub 单元测试
└── ...
```

**理由：**
- Go 社区标准做法
- 便于查找和维护
- IDE 支持好

---

## 4. Executor 测试覆盖

### 4.1 单元测试

| 测试用例 | 说明 |
|---------|------|
| `TestValidateCommand` | 命令验证逻辑（合法/非法命令） |
| `TestRunTask_Success` | 任务执行成功 |
| `TestRunTask_Timeout` | 任务超时 |
| `TestRunTask_Cancel` | 任务取消 |
| `TestCancelTask` | 取消单个任务 |
| `TestCancelRun` | 取消整个 RunID |
| `TestRunStageTasksWithRunID` | DAG 阶段执行 |

### 4.2 Mock 依赖

使用接口 + manual mock：

```go
// Mock TaskRepository
type mockTaskRepo struct {
    tasks       map[int]*domain.Task
    saveCalled  bool
}

func (m *mockTaskRepo) GetByID(tid int) (*domain.Task, error) {
    return m.tasks[tid], nil
}

func (m *mockTaskRepo) Save(t *domain.Task) error {
    m.saveCalled = true
    return nil
}
```

### 4.3 测试命令

```bash
# 运行所有测试
go test ./...

# 运行单个包
go test -v ./internal/service/

# 运行特定测试
go test -v -run "TestExecutor" ./internal/service/
```

---

## 5. StreamHub 测试覆盖

| 测试用例 | 说明 |
|---------|------|
| `TestSubscribe` | 订阅与取消订阅 |
| `TestPublish` | 消息广播 |
| `TestBackpressure` | 慢客户端断开（背压测试） |

---

## 6. 实施步骤

1. 添加 `testify/assert` 依赖
2. 为 `Executor` 编写单元测试
3. 为 `StreamHub` 编写单元测试
4. 验证所有测试通过：`go test ./...`
5. 可选：集成到 CI/CD

---

## 7. 优先级

1. **Executor** — 核心任务执行器，涉及 DAG 执行、取消、超时等复杂逻辑
2. **StreamHub** — SSE 消息广播，相对独立，易于测试
