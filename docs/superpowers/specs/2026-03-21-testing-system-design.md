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
| `TestValidateCommand_DangerousChars` | 危险字符检测（; & $ ` \\ | < >） |
| `TestRunTask_Success` | 任务执行成功 |
| `TestRunTask_Timeout` | 任务超时 |
| `TestRunTask_Cancel` | 任务取消 |
| `TestRunTask_EmptyCommand` | 空命令处理 |
| `TestRunTask_Directory` | 指定工作目录 |
| `TestCancelTask` | 取消单个任务 |
| `TestCancelRun` | 取消整个 RunID |
| `TestCancelRun_CancelledTasks` | 取消后后续任务不再执行 |
| `TestRunStageTasksWithRunID_Basic` | DAG 阶段执行（基本场景） |
| `TestRunStageTasksWithRunID_Empty` | 空 DAG 处理 |
| `TestRunStageTasksWithRunID_SingleNode` | 单节点 DAG |
| `TestRunStageTasksWithRunID_Concurrent` | 同阶段多任务并发执行 |
| `TestIsContainerRunning` | 容器运行状态查询 |

### 4.2 Mock 依赖

使用接口 + manual mock，所有依赖均通过接口注入：

```go
// mockTaskRepo implements TaskRepository interface
type mockTaskRepo struct {
    tasks      map[int]*domain.Task
    saveCalled bool
}

func (m *mockTaskRepo) GetByID(tid int) (*domain.Task, error) {
    if t, ok := m.tasks[tid]; ok {
        return t, nil
    }
    return nil, errors.New("task not found")
}

func (m *mockTaskRepo) Save(t *domain.Task) error {
    m.saveCalled = true
    if m.tasks == nil {
        m.tasks = make(map[int]*domain.Task)
    }
    m.tasks[t.Tid] = t
    return nil
}

// mockRelationRepo implements RelationRepository interface
type mockRelationRepo struct {
    relations []*domain.Relation
}

func (m *mockRelationRepo) GetByCID(cid int) ([]*domain.Relation, error) {
    return m.relations, nil
}

// mockTaskLogRepo implements TaskLogRepository interface
type mockTaskLogRepo struct {
    logs []*domain.TaskLog
}

func (m *mockTaskLogRepo) Save(log *domain.TaskLog) error {
    m.logs = append(m.logs, log)
    return nil
}

// mockContainerRepo implements ContainerRepository interface
type mockContainerRepo struct {
    containers map[int]*domain.Container
}

func (m *mockContainerRepo) GetByID(cid int) (*domain.Container, error) {
    if c, ok := m.containers[cid]; ok {
        return c, nil
    }
    return nil, errors.New("container not found")
}

func (m *mockContainerRepo) Save(c *domain.Container) error {
    m.containers[c.Cid] = c
    return nil
}
```

### 4.3 测试策略

- **单元测试**：每个接口使用 mock，不依赖真实数据库
- **边界场景**：空 DAG、单节点 DAG、多起点 DAG
- **并发安全**：同阶段任务并发执行的互斥保护测试
- **错误处理**：数据库错误、命令执行失败、超时场景

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
| `TestSubscribe_ContextCancel` | Context 取消时自动移除订阅 |
| `TestPublish` | 消息广播到所有订阅者 |
| `TestPublish_SingleSubscriber` | 单订阅者接收 |
| `TestPublish_MultipleSubscribers` | 多订阅者同时接收 |
| `TestBackpressure_SlowClient` | 慢客户端 channel 满时被断开 |
| `TestBackpressure_RemoveOnFull` | channel 满时正确移除订阅 |
| `TestPublish_IDAndTS` | 消息 ID 和时间戳自动填充 |

### 5.1 背压测试说明

`StreamHub` 的背压策略：若订阅者 channel 满（缓冲区已满），则断开该订阅者。

```go
func TestBackpressure_SlowClient(t *testing.T) {
    hub := NewStreamHub(2) // 缓冲区大小为 2
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    ch := hub.Subscribe(ctx)

    // 发送 3 条消息（缓冲区为 2）
    hub.Publish(StreamEvent{Msg: "msg1"})
    hub.Publish(StreamEvent{Msg: "msg2"})
    hub.Publish(StreamEvent{Msg: "msg3"}) // 这条会导致慢客户端被断开

    // 验证 channel 已被关闭
    _, ok := <-ch
    assert.False(t, ok, "slow client should be disconnected")
}
```

---

## 6. 实施步骤

1. 添加 `testify/assert` 依赖：`go get github.com/stretchr/testify`
2. 为 `Executor` 编写单元测试（`executor_test.go`）
3. 为 `StreamHub` 编写单元测试（`stream_hub_test.go`）
4. 验证所有测试通过：`go test ./...`
5. 可选：集成到 CI/CD

### 6.1 CI/CD 集成（可选）

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run tests
        run: go test -v -race ./...
```

**注意：** 本次实施聚焦于 Executor 和 StreamHub 的单元测试。Handler、Repository、SchedulerService 等模块的测试可在后续迭代中添加。

---

## 7. 测试范围与优先级

### 本次实施范围（v1）

1. **Executor** — 核心任务执行器，涉及 DAG 执行、取消、超时等复杂逻辑
2. **StreamHub** — SSE 消息广播，相对独立，易于测试

### 未来迭代（v2+）

- Repository 层测试（可使用 SQLite 内存数据库）
- Handler 层测试（HTTP 端点测试）
- SchedulerService 测试
- 中间件测试

### 测试金字塔

```
        /\
       /  \      E2E (少量)
      /----\
     /      \    Integration (中等)
    /--------\
   /          \  Unit (大量)
  /____________\
```

本次实施聚焦 **单元测试** 层，为后续集成测试打好基础。
