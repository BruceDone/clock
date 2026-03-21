# 测试体系实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**目标：** 为 Clock 项目建立测试体系，覆盖 Executor 和 StreamHub 核心模块

**架构：** 采用分散式测试结构（测试文件与被测文件同目录），使用 testify/assert 断言库，通过接口 + manual mock 隔离依赖

**技术栈：** Go 标准库 `testing` + `github.com/stretchr/testify/assert`

---

## 文件结构

```
internal/service/
├── executor.go           # 被测文件
├── executor_test.go      # 新建：Executor 单元测试
├── stream_hub.go         # 被测文件
├── stream_hub_test.go    # 新建：StreamHub 单元测试
├── interfaces.go         # 已有：服务接口定义
└── ...

internal/repository/
└── repository.go         # 已有：仓储接口定义
```

---

## Task 1: 安装 testify 依赖

- [ ] **Step 1: 安装 testify**

```bash
go get github.com/stretchr/testify
```

---

## Task 2: 创建 StreamHub 测试

**Files:**
- 创建: `internal/service/stream_hub_test.go`
- 参考: `internal/service/stream_hub.go`

- [ ] **Step 1: 创建测试文件框架**

```go
package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSubscribe(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	// 验证 channel 存在
	assert.NotNil(t, ch)

	// 清理
	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestSubscribe_ContextCancel(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())

	ch := hub.Subscribe(ctx)
	cancel()
	time.Sleep(10 * time.Millisecond)

	// 验证 channel 已关闭
	_, ok := <-ch
	assert.False(t, ok, "channel should be closed after context cancel")
}

func TestPublish(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	hub.Publish(StreamEvent{
		Kind: "test",
		Msg:  "hello",
	})

	// 验证收到消息
	select {
	case ev := <-ch:
		assert.Equal(t, "test", ev.Kind)
		assert.Equal(t, "hello", ev.Msg)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestPublish_SingleSubscriber(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	hub.Publish(StreamEvent{Kind: "stdout", Msg: "line1"})
	hub.Publish(StreamEvent{Kind: "stdout", Msg: "line2"})

	ev1 := <-ch
	ev2 := <-ch

	assert.Equal(t, "line1", ev1.Msg)
	assert.Equal(t, "line2", ev2.Msg)
}

func TestPublish_MultipleSubscribers(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1 := hub.Subscribe(ctx)
	ch2 := hub.Subscribe(ctx)

	hub.Publish(StreamEvent{Kind: "test", Msg: "broadcast"})

	ev1 := <-ch1
	ev2 := <-ch2

	assert.Equal(t, "broadcast", ev1.Msg)
	assert.Equal(t, "broadcast", ev2.Msg)
}

func TestBackpressure_SlowClient(t *testing.T) {
	hub := NewStreamHub(2) // 缓冲区大小为 2
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	// 发送 3 条消息（缓冲区为 2），第三条会导致断开
	hub.Publish(StreamEvent{Msg: "msg1"})
	hub.Publish(StreamEvent{Msg: "msg2"})
	hub.Publish(StreamEvent{Msg: "msg3"}")

	// 验证 channel 已被关闭
	_, ok := <-ch
	assert.False(t, ok, "slow client should be disconnected")
}

func TestPublish_IDAndTS(t *testing.T) {
	hub := NewStreamHub(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := hub.Subscribe(ctx)

	hub.Publish(StreamEvent{Kind: "test", Msg: "msg"})

	select {
	case ev := <-ch:
		assert.True(t, ev.ID > 0, "ID should be auto-assigned")
		assert.True(t, ev.TS > 0, "TS should be auto-assigned")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}
```

- [ ] **Step 2: 运行 StreamHub 测试验证**

```bash
go test -v -run "TestStreamHub" ./internal/service/
```

Expected: PASS (7 tests)

- [ ] **Step 3: 提交**

```bash
git add internal/service/stream_hub_test.go
git commit -m "test: add StreamHub unit tests"
```

---

## Task 3: 创建 Executor 测试 - 第一部分（验证与基础运行）

**Files:**
- 创建: `internal/service/executor_test.go`
- 参考: `internal/service/executor.go`, `internal/domain/task.go`

- [ ] **Step 1: 创建测试文件框架和 Mock**

```go
package service

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"clock/internal/domain"
)

// mockTaskRepo implements TaskRepository for testing
type mockTaskRepo struct {
	mu      sync.Mutex
	tasks   map[int]*domain.Task
	saveErr error
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{
		tasks: make(map[int]*domain.Task),
	}
}

func (m *mockTaskRepo) GetByID(tid int) (*domain.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tasks[tid]; ok {
		return t, nil
	}
	return nil, errors.New("task not found")
}

func (m *mockTaskRepo) Save(task *domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	m.tasks[task.Tid] = task
	return nil
}

func (m *mockTaskRepo) List(query interface{}) ([]*domain.Task, error) {
	return nil, nil
}

func (m *mockTaskRepo) GetByCID(cid int) ([]*domain.Task, error) {
	return nil, nil
}

func (m *mockTaskRepo) Delete(tid int) error {
	return nil
}

func (m *mockTaskRepo) DeleteByCID(cid int) error {
	return nil
}

func (m *mockTaskRepo) UpdateCoordinates(tid, x, y int) error {
	return nil
}

// mockRelationRepo implements RelationRepository for testing
type mockRelationRepo struct {
	relations []*domain.Relation
}

func (m *mockRelationRepo) GetByCID(cid int) ([]*domain.Relation, error) {
	return m.relations, nil
}

func (m *mockRelationRepo) Save(relation *domain.Relation) error {
	return nil
}

func (m *mockRelationRepo) Delete(rid int) error {
	return nil
}

func (m *mockRelationRepo) DeleteByTID(tid int) error {
	return nil
}

func (m *mockRelationRepo) DeleteByNextTID(nextTid int) error {
	return nil
}

// mockTaskLogRepo implements TaskLogRepository for testing
type mockTaskLogRepo struct {
	logs []*domain.TaskLog
}

func (m *mockTaskLogRepo) List(query interface{}) ([]*domain.TaskLog, error) {
	return nil, nil
}

func (m *mockTaskLogRepo) Save(log *domain.TaskLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockTaskLogRepo) DeleteByID(lid string) error {
	return nil
}

func (m *mockTaskLogRepo) DeleteByTimeRange(query interface{}) error {
	return nil
}

func (m *mockTaskLogRepo) DeleteAll() error {
	return nil
}

// mockContainerRepo implements ContainerRepository for testing
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
	if m.containers == nil {
		m.containers = make(map[int]*domain.Container)
	}
	m.containers[c.Cid] = c
	return nil
}

func (m *mockContainerRepo) List(query interface{}) ([]*domain.Container, error) {
	return nil, nil
}

func (m *mockContainerRepo) FindAll() ([]*domain.Container, error) {
	return nil, nil
}

func (m *mockContainerRepo) Delete(cid int) error {
	return nil
}

// 创建测试任务辅助函数
func newTestTask(tid, cid int, command string) *domain.Task {
	return &domain.Task{
		Tid:       tid,
		Cid:       cid,
		Command:   command,
		Name:      "test-task-" + string(rune(tid+'0')),
		Status:    domain.StatusPending,
		Timeout:   30,
		LogEnable: false,
	}
}
```

- [ ] **Step 2: 添加命令验证测试**

```go
func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		wantErr bool
	}{
		{"valid_simple", "echo hello", false},
		{"valid_with_args", "ls -la /tmp", false},
		{"valid_with_path", "/usr/bin/ls", false},
		{"valid_special_chars", "echo hello world", false},
		{"empty", "", true},
		{"whitespace_only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(tt.cmd)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCommand_DangerousChars(t *testing.T) {
	dangerous := []string{
		"echo hello; rm -rf /",
		"echo hello & whoami",
		"echo `$HOME`",
		"echo $(whoami)",
		"cat /etc/passwd | grep root",
		"echo hello > file",
		"echo 'hello'",
		`echo "hello"`,
	}

	for _, cmd := range dangerous {
		t.Run(cmd, func(t *testing.T) {
			err := validateCommand(cmd)
			assert.Error(t, err, "command should be rejected: %s", cmd)
		})
	}
}
```

- [ ] **Step 3: 添加 RunTask 基础测试**

```go
func TestRunTask_Success(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	task := newTestTask(1, 1, "echo hello")
	taskRepo.tasks[1] = task

	err := executor.RunTask(task)

	assert.NoError(t, err)
	assert.Equal(t, domain.StatusSuccess, task.Status)
}

func TestRunTask_EmptyCommand(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	task := newTestTask(1, 1, "")
	taskRepo.tasks[1] = task

	err := executor.RunTask(task)

	assert.Error(t, err)
	assert.Equal(t, domain.StatusFailure, task.Status)
	assert.Contains(t, err.Error(), "empty")
}
```

- [ ] **Step 4: 运行测试验证**

```bash
go test -v -run "TestValidateCommand|TestRunTask_Success|TestRunTask_EmptyCommand" ./internal/service/
```

Expected: PASS (5 tests)

- [ ] **Step 5: 提交**

```bash
git add internal/service/executor_test.go
git commit -m "test: add Executor command validation and basic run tests"
```

---

## Task 4: 创建 Executor 测试 - 第二部分（超时与取消）

**Files:**
- 修改: `internal/service/executor_test.go`

- [ ] **Step 1: 添加超时和取消测试**

```go
func TestRunTask_Timeout(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	task := newTestTask(1, 1, "sleep 10")
	task.Timeout = 1 // 1 second timeout
	taskRepo.tasks[1] = task

	err := executor.RunTask(task)

	assert.Error(t, err)
	assert.Equal(t, domain.StatusFailure, task.Status)
	assert.Contains(t, err.Error(), "timeout")
}

func TestRunTask_Cancel(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	task := newTestTask(1, 1, "sleep 60")
	task.Timeout = 0 // no timeout
	taskRepo.tasks[1] = task

	// 在 goroutine 中运行任务，然后取消
	done := make(chan error, 1)
	go func() {
		done <- executor.RunTask(task)
	}()

	time.Sleep(100 * time.Millisecond)
	err := executor.CancelTask(1)

	// 取消不返回错误（取消操作本身成功）
	assert.NoError(t, err)

	// 等待任务结束
	select {
	case err := <-done:
		assert.Error(t, err)
		assert.Equal(t, domain.StatusCancelled, task.Status)
	case <-time.After(2 * time.Second):
		t.Fatal("task did not finish in time")
	}
}

func TestCancelTask_NotRunning(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	err := executor.CancelTask(999) // 不存在的任务

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestCancelRun(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	// 取消空的 runID
	err := executor.CancelRun("")
	assert.Error(t, err)

	// 取消不存在的 runID（应该成功，只是没有任务可取消）
	err = executor.CancelRun("nonexistent")
	assert.NoError(t, err)
}
```

- [ ] **Step 2: 运行测试验证**

```bash
go test -v -run "TestRunTask_Timeout|TestRunTask_Cancel|TestCancelTask|TestCancelRun" ./internal/service/
```

Expected: PASS (5 tests)

- [ ] **Step 3: 提交**

```bash
git add internal/service/executor_test.go
git commit -m "test: add Executor timeout and cancel tests"
```

---

## Task 5: 创建 Executor 测试 - 第三部分（DAG 执行）

**Files:**
- 修改: `internal/service/executor_test.go`

- [ ] **Step 1: 添加 DAG 执行测试**

```go
func TestRunStageTasksWithRunID_Empty(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	// 空任务列表不应 panic
	executor.runStageTasksWithRunID(nil, nil, "test-run")

	// 验证无 panic，测试通过
	assert.True(t, true)
}

func TestRunStageTasksWithRunID_SingleNode(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{relations: []*domain.Relation{}}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	task := newTestTask(1, 1, "echo single")
	taskRepo.tasks[1] = task

	executor.runStageTasksWithRunID([]*domain.Task{task}, nil, "test-run")

	// 单节点 DAG 应该执行完成
	assert.Equal(t, domain.StatusSuccess, task.Status)
}

func TestRunStageTasksWithRunID_TwoStage(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{
		relations: []*domain.Relation{
			{Rid: 1, Tid: 1, NextTid: 2}, // Task1 -> Task2
		},
	}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	task1 := newTestTask(1, 1, "echo first")
	task2 := newTestTask(2, 1, "echo second")
	taskRepo.tasks[1] = task1
	taskRepo.tasks[2] = task2

	executor.runStageTasksWithRunID([]*domain.Task{task1, task2}, relRepo.relations, "test-run")

	assert.Equal(t, domain.StatusSuccess, task1.Status)
	assert.Equal(t, domain.StatusSuccess, task2.Status)
}

func TestIsContainerRunning(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	// 初始状态：容器未运行
	assert.False(t, executor.IsContainerRunning(1))

	// 注册运行中的容器（通过内部方法，这里测试其逻辑）
	executor.registerRunningContainer(1, "run-123")
	assert.True(t, executor.IsContainerRunning(1))

	executor.unregisterRunningContainer(1)
	assert.False(t, executor.IsContainerRunning(1))
}
```

- [ ] **Step 2: 运行测试验证**

```bash
go test -v -run "TestRunStageTasks|TestIsContainerRunning" ./internal/service/
```

Expected: PASS (5 tests)

- [ ] **Step 3: 提交**

```bash
git add internal/service/executor_test.go
git commit -m "test: add Executor DAG execution tests"
```

---

## Task 6: 最终验证

- [ ] **Step 1: 运行所有测试**

```bash
go test -v ./internal/service/
```

Expected: PASS (all ~20 tests)

- [ ] **Step 2: 运行完整测试套件**

```bash
go test ./...
```

- [ ] **Step 3: 运行 go vet 检查**

```bash
go vet ./...
```

---

## 任务完成检查清单

- [ ] testify 依赖已安装
- [ ] StreamHub 测试 (7 tests) 全部通过
- [ ] Executor 测试 (~15 tests) 全部通过
- [ ] `go test ./...` 全部通过
- [ ] `go vet ./...` 无警告
- [ ] 所有更改已提交
