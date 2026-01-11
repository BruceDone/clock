package service

import (
	"bufio"
	"bytes"
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

// Executor 任务执行器
type Executor struct {
	taskRepo     repository.TaskRepository
	relationRepo repository.RelationRepository
	taskLogRepo  repository.TaskLogRepository
	hub          *StreamHub
}

// NewExecutor 创建执行器
func NewExecutor(
	taskRepo repository.TaskRepository,
	relationRepo repository.RelationRepository,
	taskLogRepo repository.TaskLogRepository,
	hub *StreamHub,
) *Executor {
	return &Executor{
		taskRepo:     taskRepo,
		relationRepo: relationRepo,
		taskLogRepo:  taskLogRepo,
		hub:          hub,
	}
}

// RunTask 执行单个任务
func (e *Executor) RunTask(task *domain.Task) error {
	var stdOutBuf bytes.Buffer
	var stdErrBuf bytes.Buffer

	startAt := time.Now()
	errMsg := ""

	// 设置开始状态
	task.Status = domain.StatusStart
	defer func() {
		task.UpdateAt = time.Now().Unix()
		logger.Debugf("[%d] finished task [%s]", task.Tid, task.Name)
		_ = e.taskRepo.Save(task)
		e.saveLog(task, stdOutBuf, stdErrBuf)

		e.hub.Publish(StreamEvent{
			Kind:       "task_end",
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
		Tid:      task.Tid,
		Cid:      task.Cid,
		TaskName: task.Name,
		Msg:      "running",
	})

	// 解析命令
	args := strings.Split(task.Command, " ")
	cmd := exec.Command(args[0], args[1:]...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		task.Status = domain.StatusFailure
		errMsg = err.Error()
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		task.Status = domain.StatusFailure
		errMsg = err.Error()
		return err
	}

	if task.Directory != "" {
		cmd.Dir = task.Directory
	}

	if err := cmd.Start(); err != nil {
		task.Status = domain.StatusFailure
		errMsg = err.Error()
		return err
	}

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
		case <-time.After(time.Duration(task.Timeout) * time.Second):
			timedOut = true
			_ = cmd.Process.Kill()
			waitErr = <-done
		}
	} else {
		waitErr = cmd.Wait()
	}

	// Ensure readers have drained pipes before we persist logs.
	wg.Wait()

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

// RunTaskByID 根据ID执行任务（检查依赖）
func (e *Executor) RunTaskByID(tid int) error {
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
			// 如果有前置任务失败或等待中，当前任务设为等待
			if preTask.Status == domain.StatusFailure || preTask.Status == domain.StatusPending {
				task.Status = domain.StatusPending
				_ = e.taskRepo.Save(task)
				return nil
			}
		}
	}

	return e.RunTask(task)
}

// RunContainer 按DAG拓扑顺序执行容器内所有任务
func (e *Executor) RunContainer(container *domain.Container, tasks []*domain.Task, relations []*domain.Relation) error {
	// 设置容器开始状态
	container.Status = domain.StatusStart
	// 注意：这里不能直接保存，因为executor没有containerRepo

	defer func() {
		container.Status = domain.StatusSuccess
	}()

	e.runStageTasks(tasks, relations)
	return nil
}

// runStageTasks 按阶段执行任务（拓扑排序）
func (e *Executor) runStageTasks(tasks []*domain.Task, relations []*domain.Relation) {
	stage := 0

	// 复制任务和关系列表
	taskList := make([]*domain.Task, len(tasks))
	copy(taskList, tasks)

	relationList := make([]*domain.Relation, len(relations))
	copy(relationList, relations)

	for {
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
			if err := e.RunTaskByID(tid); err != nil {
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
