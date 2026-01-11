package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	uuid "github.com/nu7hatch/gouuid"

	"clock/internal/domain"
	"clock/internal/logger"
	"clock/internal/repository"
	"clock/pkg/util"
)

// runningTask 运行中的任务信息
type runningTask struct {
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	runID    string
	tid      int
	cid      int
	taskName string
	startAt  time.Time
}

// RunningTaskInfo 运行中任务信息（用于返回给前端）
type RunningTaskInfo struct {
	Tid      int    `json:"tid"`
	Cid      int    `json:"cid"`
	RunID    string `json:"runId"`
	TaskName string `json:"taskName"`
	StartAt  int64  `json:"startAt"`
}

// Executor 任务执行器
type Executor struct {
	taskRepo      repository.TaskRepository
	relationRepo  repository.RelationRepository
	taskLogRepo   repository.TaskLogRepository
	containerRepo repository.ContainerRepository
	hub           *StreamHub

	runningMu sync.RWMutex
	running   map[int]*runningTask // key: tid

	// cancelledRuns 用于标记已取消的 runID，防止 DAG 中后续任务继续执行
	cancelledRunsMu sync.RWMutex
	cancelledRuns   map[string]struct{}

	// runningContainers 用于追踪正在运行的容器，实现阻塞式运行
	runningContainersMu sync.RWMutex
	runningContainers   map[int]string // key: cid, value: runID
}

// NewExecutor 创建执行器
func NewExecutor(
	taskRepo repository.TaskRepository,
	relationRepo repository.RelationRepository,
	taskLogRepo repository.TaskLogRepository,
	containerRepo repository.ContainerRepository,
	hub *StreamHub,
) *Executor {
	return &Executor{
		taskRepo:          taskRepo,
		relationRepo:      relationRepo,
		taskLogRepo:       taskLogRepo,
		containerRepo:     containerRepo,
		hub:               hub,
		running:           make(map[int]*runningTask),
		cancelledRuns:     make(map[string]struct{}),
		runningContainers: make(map[int]string),
	}
}

// RunTask 执行单个任务
func (e *Executor) RunTask(task *domain.Task) error {
	return e.RunTaskWithRunID(task, "")
}

// RunTaskWithRunID 执行单个任务，带 runID 用于日志追踪
func (e *Executor) RunTaskWithRunID(task *domain.Task, runID string) error {
	// 检查该 runID 是否已被取消
	if runID != "" {
		e.cancelledRunsMu.RLock()
		_, cancelled := e.cancelledRuns[runID]
		e.cancelledRunsMu.RUnlock()
		if cancelled {
			task.Status = domain.StatusCancelled
			task.UpdateAt = time.Now().Unix()
			_ = e.taskRepo.Save(task)
			e.hub.Publish(StreamEvent{
				Kind:     "task_end",
				RunID:    runID,
				Tid:      task.Tid,
				Cid:      task.Cid,
				TaskName: task.Name,
				Status:   "cancelled",
				Msg:      "run was cancelled",
			})
			return errors.New("run was cancelled")
		}
	}

	var stdOutBuf bytes.Buffer
	var stdErrBuf bytes.Buffer

	startAt := time.Now()
	errMsg := ""
	cancelled := false

	// 创建可取消的 context
	ctx, cancel := context.WithCancel(context.Background())

	// 设置开始状态
	task.Status = domain.StatusStart
	defer func() {
		// 从 running map 中移除
		e.runningMu.Lock()
		delete(e.running, task.Tid)
		e.runningMu.Unlock()

		task.UpdateAt = time.Now().Unix()
		logger.Debugf("[%d] finished task [%s]", task.Tid, task.Name)
		_ = e.taskRepo.Save(task)
		e.saveLog(task, stdOutBuf, stdErrBuf)

		e.hub.Publish(StreamEvent{
			Kind:       "task_end",
			RunID:      runID,
			Tid:        task.Tid,
			Cid:        task.Cid,
			TaskName:   task.Name,
			Status:     domain.StatusText(task.Status),
			DurationMs: time.Since(startAt).Milliseconds(),
			Msg:        errMsg,
		})
	}()

	if task.Command == "" {
		task.Status = domain.StatusFailure
		errMsg = "command cannot be empty"
		return errors.New(errMsg)
	}

	logger.Debugf("[%d] running task [%s]", task.Tid, task.Name)
	_ = e.taskRepo.Save(task)

	e.hub.Publish(StreamEvent{
		Kind:     "task_start",
		RunID:    runID,
		Tid:      task.Tid,
		Cid:      task.Cid,
		TaskName: task.Name,
		Msg:      "running",
	})

	// 解析命令，使用 CommandContext 支持取消
	args := strings.Split(task.Command, " ")
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		task.Status = domain.StatusFailure
		errMsg = err.Error()
		cancel()
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		task.Status = domain.StatusFailure
		errMsg = err.Error()
		cancel()
		return err
	}

	if task.Directory != "" {
		cmd.Dir = task.Directory
	}

	if err := cmd.Start(); err != nil {
		task.Status = domain.StatusFailure
		errMsg = err.Error()
		cancel()
		return err
	}

	// 注册到 running map
	e.runningMu.Lock()
	e.running[task.Tid] = &runningTask{
		cmd:      cmd,
		cancel:   cancel,
		runID:    runID,
		tid:      task.Tid,
		cid:      task.Cid,
		taskName: task.Name,
		startAt:  startAt,
	}
	e.runningMu.Unlock()

	var wg sync.WaitGroup
	wg.Add(2)

	readStream := func(kind string, r *bufio.Scanner, buf *bytes.Buffer) {
		defer wg.Done()
		for r.Scan() {
			line := r.Text()
			buf.WriteString(line)
			buf.WriteByte('\n')
			e.hub.Publish(StreamEvent{
				Kind:     kind,
				RunID:    runID,
				Tid:      task.Tid,
				Cid:      task.Cid,
				TaskName: task.Name,
				Msg:      line,
			})
		}
		if scanErr := r.Err(); scanErr != nil {
			buf.WriteString(scanErr.Error())
			buf.WriteByte('\n')
			e.hub.Publish(StreamEvent{
				Kind:     "meta",
				RunID:    runID,
				Tid:      task.Tid,
				Cid:      task.Cid,
				TaskName: task.Name,
				Msg:      fmt.Sprintf("%s reader error: %v", kind, scanErr),
			})
		}
	}

	stdoutScanner := bufio.NewScanner(stdoutPipe)
	stdoutScanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	stderrScanner := bufio.NewScanner(stderrPipe)
	stderrScanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	go readStream("stdout", stdoutScanner, &stdOutBuf)
	go readStream("stderr", stderrScanner, &stdErrBuf)

	// Wait for process completion (with optional timeout)
	var waitErr error
	var timedOut bool

	if task.Timeout > 0 {
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case waitErr = <-done:
			// 进程已结束，检查是否是被取消导致的
			// (CommandContext 在 ctx 取消时会 kill 进程)
		case <-time.After(time.Duration(task.Timeout) * time.Second):
			timedOut = true
			_ = cmd.Process.Kill()
			waitErr = <-done
		case <-ctx.Done():
			// 被取消
			cancelled = true
			// CommandContext 已经会 kill 进程，但为了确保，再次调用
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			waitErr = <-done
		}
	} else {
		waitErr = cmd.Wait()
	}

	// 检查是否是被取消（无论是通过 ctx.Done() 还是进程自然结束时 ctx 已被取消）
	if ctx.Err() == context.Canceled {
		cancelled = true
	}

	// Ensure readers have drained pipes before we persist logs.
	wg.Wait()

	// 优先判断取消状态（取消优先于超时）
	if cancelled {
		task.Status = domain.StatusCancelled
		errMsg = "task cancelled by user"
		logger.Infof("task %s was cancelled", task.Name)
		return errors.New(errMsg)
	}

	if timedOut {
		task.Status = domain.StatusFailure
		errMsg = fmt.Sprintf("task %s timeout", task.Name)
		logger.Errorf("task %s reached timeout limit", task.Name)
		return errors.New(errMsg)
	}

	if waitErr != nil {
		task.Status = domain.StatusFailure
		stdErrBuf.WriteString(waitErr.Error())
		stdErrBuf.WriteByte('\n')
		errMsg = waitErr.Error()
		return waitErr
	}

	task.Status = domain.StatusSuccess
	return nil
}

// CancelTask 取消单个任务
func (e *Executor) CancelTask(tid int) error {
	e.runningMu.RLock()
	rt, ok := e.running[tid]
	e.runningMu.RUnlock()

	if !ok {
		return errors.New("task not running")
	}

	// 调用 cancel 函数取消任务
	rt.cancel()
	logger.Infof("task %d cancelled by user", tid)
	return nil
}

// CancelRun 取消整个 runID 关联的所有任务
func (e *Executor) CancelRun(runID string) error {
	if runID == "" {
		return errors.New("runID cannot be empty")
	}

	// 标记该 runID 为已取消，防止后续任务执行
	e.cancelledRunsMu.Lock()
	e.cancelledRuns[runID] = struct{}{}
	e.cancelledRunsMu.Unlock()

	// 取消所有属于该 runID 的运行中任务
	e.runningMu.RLock()
	var toCancel []*runningTask
	for _, rt := range e.running {
		if rt.runID == runID {
			toCancel = append(toCancel, rt)
		}
	}
	e.runningMu.RUnlock()

	for _, rt := range toCancel {
		rt.cancel()
		logger.Infof("task %d (runID: %s) cancelled by user", rt.tid, runID)
	}

	// 10分钟后清理 cancelledRuns 中的记录
	go func() {
		time.Sleep(10 * time.Minute)
		e.cancelledRunsMu.Lock()
		delete(e.cancelledRuns, runID)
		e.cancelledRunsMu.Unlock()
	}()

	return nil
}

// GetRunningTasks 获取正在运行的任务列表
func (e *Executor) GetRunningTasks() []RunningTaskInfo {
	e.runningMu.RLock()
	defer e.runningMu.RUnlock()

	result := make([]RunningTaskInfo, 0, len(e.running))
	for _, rt := range e.running {
		result = append(result, RunningTaskInfo{
			Tid:      rt.tid,
			Cid:      rt.cid,
			RunID:    rt.runID,
			TaskName: rt.taskName,
			StartAt:  rt.startAt.UnixMilli(),
		})
	}
	return result
}

// IsContainerRunning 检查容器是否正在运行
func (e *Executor) IsContainerRunning(cid int) bool {
	e.runningContainersMu.RLock()
	defer e.runningContainersMu.RUnlock()
	_, ok := e.runningContainers[cid]
	return ok
}

// registerRunningContainer 注册运行中的容器
func (e *Executor) registerRunningContainer(cid int, runID string) {
	e.runningContainersMu.Lock()
	defer e.runningContainersMu.Unlock()
	e.runningContainers[cid] = runID
}

// unregisterRunningContainer 注销运行中的容器
func (e *Executor) unregisterRunningContainer(cid int) {
	e.runningContainersMu.Lock()
	defer e.runningContainersMu.Unlock()
	delete(e.runningContainers, cid)
}

// GetRunningContainers 获取运行中的容器列表
func (e *Executor) GetRunningContainers() []int {
	e.runningContainersMu.RLock()
	defer e.runningContainersMu.RUnlock()
	result := make([]int, 0, len(e.runningContainers))
	for cid := range e.runningContainers {
		result = append(result, cid)
	}
	return result
}

// RunTaskByID 根据ID执行任务（检查依赖）
func (e *Executor) RunTaskByID(tid int) error {
	return e.RunTaskByIDWithRunID(tid, "")
}

// RunTaskByIDWithRunID 根据ID执行任务，带 runID
func (e *Executor) RunTaskByIDWithRunID(tid int, runID string) error {
	// 检查该 runID 是否已被取消
	if runID != "" {
		e.cancelledRunsMu.RLock()
		_, cancelled := e.cancelledRuns[runID]
		e.cancelledRunsMu.RUnlock()
		if cancelled {
			return errors.New("run was cancelled")
		}
	}

	task, err := e.taskRepo.GetByID(tid)
	if err != nil {
		return err
	}

	// 检查前置任务是否都已成功
	relations, err := e.relationRepo.GetByCID(task.Cid)
	if err != nil {
		return err
	}

	// 找出当前任务的前置任务
	var preTids []int
	for _, rel := range relations {
		if rel.NextTid == tid {
			preTids = append(preTids, rel.Tid)
		}
	}

	// 如果有前置任务，检查它们的状态
	if len(preTids) > 0 {
		for _, preTid := range preTids {
			preTask, err := e.taskRepo.GetByID(preTid)
			if err != nil {
				continue
			}
			// 如果有前置任务失败、等待中或已取消，当前任务设为等待
			if preTask.Status == domain.StatusFailure || preTask.Status == domain.StatusPending || preTask.Status == domain.StatusCancelled {
				task.Status = domain.StatusPending
				_ = e.taskRepo.Save(task)
				return nil
			}
		}
	}

	return e.RunTaskWithRunID(task, runID)
}

// RunContainer 按DAG拓扑顺序执行容器内所有任务
func (e *Executor) RunContainer(container *domain.Container, tasks []*domain.Task, relations []*domain.Relation) error {
	// 检查阻塞运行
	if container.Blocking {
		if e.IsContainerRunning(container.Cid) {
			logger.Infof("[executor] container %s (cid=%d) is running, skip this schedule", container.Name, container.Cid)
			return nil
		}
	}

	// 生成本次执行的唯一 runID
	runID := genGUID(8)

	// 注册运行中的容器
	e.registerRunningContainer(container.Cid, runID)
	defer e.unregisterRunningContainer(container.Cid)

	// 设置容器开始状态并保存
	container.Status = domain.StatusStart
	container.UpdateAt = time.Now().Unix()
	_ = e.containerRepo.Save(container)

	defer func() {
		container.Status = domain.StatusPending
		container.UpdateAt = time.Now().Unix()
		_ = e.containerRepo.Save(container)
	}()

	// 重置所有任务状态为 Pending，确保干净的执行环境
	for _, task := range tasks {
		task.Status = domain.StatusPending
		_ = e.taskRepo.Save(task)
	}

	e.runStageTasksWithRunID(tasks, relations, runID)
	return nil
}

// runStageTasks 按阶段执行任务（拓扑排序）
func (e *Executor) runStageTasks(tasks []*domain.Task, relations []*domain.Relation) {
	e.runStageTasksWithRunID(tasks, relations, "")
}

// runStageTasksWithRunID 按阶段执行任务，带 runID
func (e *Executor) runStageTasksWithRunID(tasks []*domain.Task, relations []*domain.Relation, runID string) {
	stage := 0

	// 复制任务和关系列表
	taskList := make([]*domain.Task, len(tasks))
	copy(taskList, tasks)

	relationList := make([]*domain.Relation, len(relations))
	copy(relationList, relations)

	for {
		// 检查该 runID 是否已被取消
		if runID != "" {
			e.cancelledRunsMu.RLock()
			_, cancelled := e.cancelledRuns[runID]
			e.cancelledRunsMu.RUnlock()
			if cancelled {
				logger.Infof("[executor] runID %s was cancelled, stopping DAG execution", runID)
				break
			}
		}

		logger.Debugf("[executor] stage %d", stage)

		if len(taskList) == 0 {
			break
		}

		var rootTids []int

		// 初始化入度
		inDegree := make(map[int]int)
		for _, task := range taskList {
			inDegree[task.Tid] = 0
		}

		// 计算入度
		for _, rel := range relationList {
			if _, ok := inDegree[rel.NextTid]; ok {
				inDegree[rel.NextTid]++
			}
		}

		// 筛选入度为0的节点（可以执行的任务）
		for tid, degree := range inDegree {
			if degree == 0 {
				rootTids = append(rootTids, tid)
			}
		}

		// 存在环，终止
		if len(rootTids) == 0 {
			logger.Warnf("[executor] circular dependency detected")
			break
		}

		// 执行当前阶段的任务
		for _, tid := range rootTids {
			if err := e.RunTaskByIDWithRunID(tid, runID); err != nil {
				logger.Errorf("[executor] task %d failed: %v", tid, err)
			}
		}

		// 移除已执行的节点
		taskList = util.Filter(taskList, func(t *domain.Task) bool {
			return !util.ContainsInt(rootTids, t.Tid)
		})

		// 移除已执行节点的出边
		relationList = util.Filter(relationList, func(r *domain.Relation) bool {
			return !util.ContainsInt(rootTids, r.Tid)
		})

		stage++
	}
}

// saveLog 保存执行日志
func (e *Executor) saveLog(task *domain.Task, stdOut, stdErr bytes.Buffer) {
	// 保存日志到数据库
	if task.LogEnable {
		lid := genGUID(8)
		log := &domain.TaskLog{
			Lid:      lid,
			Tid:      task.Tid,
			Cid:      task.Cid,
			StdOut:   stdOut.String(),
			StdErr:   stdErr.String(),
			UpdateAt: time.Now().Unix(),
		}
		_ = e.taskLogRepo.Save(log)
	}
}

// genGUID 生成UUID
func genGUID(length int) string {
	u, err := uuid.NewV4()
	if err != nil {
		return ""
	}
	guid := strings.Replace(u.String(), "-", "", -1)
	if length > len(guid) {
		length = len(guid)
	}
	return guid[0:length]
}
