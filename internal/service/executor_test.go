package service

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"clock/internal/domain"
	"clock/internal/repository"
)

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

func (m *mockTaskRepo) List(query *repository.TaskQuery) ([]*domain.Task, error) {
	return []*domain.Task{}, nil
}

func (m *mockTaskRepo) GetByCID(cid int) ([]*domain.Task, error) {
	return []*domain.Task{}, nil
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

type mockTaskLogRepo struct {
	logs []*domain.TaskLog
}

func (m *mockTaskLogRepo) List(query *repository.LogQuery) ([]*domain.TaskLog, error) {
	return []*domain.TaskLog{}, nil
}

func (m *mockTaskLogRepo) Save(log *domain.TaskLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockTaskLogRepo) DeleteByID(lid string) error {
	return nil
}

func (m *mockTaskLogRepo) DeleteByTimeRange(query *repository.LogQuery) error {
	return nil
}

func (m *mockTaskLogRepo) DeleteAll() error {
	return nil
}

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

func (m *mockContainerRepo) List(query *repository.ContainerQuery) ([]*domain.Container, error) {
	return []*domain.Container{}, nil
}

func (m *mockContainerRepo) FindAll() ([]*domain.Container, error) {
	return []*domain.Container{}, nil
}

func (m *mockContainerRepo) Delete(cid int) error {
	return nil
}

func newTestTask(tid, cid int, command string) *domain.Task {
	return &domain.Task{
		Tid:       tid,
		Cid:       cid,
		Command:   command,
		Name:      fmt.Sprintf("test-task-%d", tid),
		Status:    domain.StatusPending,
		Timeout:   30,
		LogEnable: false,
	}
}

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
		"echo <file",
		"echo | ls",
	}

	for _, cmd := range dangerous {
		t.Run(cmd, func(t *testing.T) {
			err := validateCommand(cmd)
			assert.Error(t, err, "command should be rejected: %s", cmd)
		})
	}
}

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
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestRunTask_Directory(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	task := newTestTask(1, 1, "pwd")
	task.Directory = "/tmp"
	taskRepo.tasks[1] = task

	err := executor.RunTask(task)

	assert.NoError(t, err)
	assert.Equal(t, domain.StatusSuccess, task.Status)
}

func TestRunStageTasksWithRunID_Empty(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	executor.runStageTasksWithRunID(nil, nil, "test-run")

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

	assert.Equal(t, domain.StatusSuccess, task.Status)
}

func TestRunStageTasksWithRunID_TwoStage(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{
		relations: []*domain.Relation{
			{Rid: 1, Tid: 1, NextTid: 2},
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

	assert.False(t, executor.IsContainerRunning(1))

	executor.registerRunningContainer(1, "run-123")
	assert.True(t, executor.IsContainerRunning(1))

	executor.unregisterRunningContainer(1)
	assert.False(t, executor.IsContainerRunning(1))
}

func TestRunStageTasksWithRunID_Concurrent(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{
		relations: []*domain.Relation{
			{Rid: 1, Tid: 1, NextTid: 3},
			{Rid: 2, Tid: 2, NextTid: 3},
		},
	}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	task1 := newTestTask(1, 1, "echo first")
	task2 := newTestTask(2, 1, "echo second")
	task3 := newTestTask(3, 1, "echo third")
	taskRepo.tasks[1] = task1
	taskRepo.tasks[2] = task2
	taskRepo.tasks[3] = task3

	executor.runStageTasksWithRunID(
		[]*domain.Task{task1, task2, task3},
		relRepo.relations,
		"test-concurrent",
	)

	assert.Equal(t, domain.StatusSuccess, task1.Status)
	assert.Equal(t, domain.StatusSuccess, task2.Status)
	assert.Equal(t, domain.StatusSuccess, task3.Status)
}

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

	taskStarted := make(chan struct{})

	go func() {
		executor.RunTask(task)
		close(taskStarted)
	}()

	for i := 0; i < 50; i++ {
		time.Sleep(20 * time.Millisecond)
		executor.runningMu.RLock()
		_, exists := executor.running[1]
		executor.runningMu.RUnlock()
		if exists {
			break
		}
	}

	err := executor.CancelTask(1)
	assert.NoError(t, err)

	select {
	case <-taskStarted:
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

	err := executor.CancelTask(999)

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

	err := executor.CancelRun("")
	assert.Error(t, err)

	err = executor.CancelRun("nonexistent")
	assert.NoError(t, err)
}

func TestCancelRun_CancelledTasks(t *testing.T) {
	taskRepo := newMockTaskRepo()
	relRepo := &mockRelationRepo{}
	logRepo := &mockTaskLogRepo{}
	containerRepo := &mockContainerRepo{}
	hub := NewStreamHub(100)

	executor := NewExecutor(taskRepo, relRepo, logRepo, containerRepo, hub)

	task := newTestTask(1, 1, "sleep 30")
	taskRepo.tasks[1] = task

	go func() {
		executor.RunTaskWithRunID(task, "test-run-cancel")
	}()

	time.Sleep(50 * time.Millisecond)

	executor.CancelRun("test-run-cancel")

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, domain.StatusCancelled, task.Status)
}
