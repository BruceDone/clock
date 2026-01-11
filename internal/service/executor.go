package service

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
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
	message      *domain.Message
}

// NewExecutor 创建执行器
func NewExecutor(
	taskRepo repository.TaskRepository,
	relationRepo repository.RelationRepository,
	taskLogRepo repository.TaskLogRepository,
	message *domain.Message,
) *Executor {
	return &Executor{
		taskRepo:     taskRepo,
		relationRepo: relationRepo,
		taskLogRepo:  taskLogRepo,
		message:      message,
	}
}

// RunTask 执行单个任务
func (e *Executor) RunTask(task *domain.Task) error {
	var stdOutBuf bytes.Buffer
	var stdErrBuf bytes.Buffer

	// 设置开始状态
	task.Status = domain.StatusStart
	defer func() {
		task.UpdateAt = time.Now().Unix()
		logger.Debugf("[%d] finished task [%s]", task.Tid, task.Name)
		_ = e.taskRepo.Save(task)
		e.saveLog(task, stdOutBuf, stdErrBuf)
	}()

	if task.Command == "" {
		task.Status = domain.StatusFailure
		return errors.New("command cannot be empty")
	}

	logger.Debugf("[%d] running task [%s]", task.Tid, task.Name)
	_ = e.taskRepo.Save(task)

	// 解析命令
	args := strings.Split(task.Command, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &stdOutBuf
	cmd.Stderr = &stdErrBuf

	if task.Directory != "" {
		cmd.Dir = task.Directory
	}

	// 带超时执行
	if task.Timeout > 0 {
		timeout := time.After(time.Duration(task.Timeout) * time.Second)
		done := make(chan error, 1)

		go func() {
			done <- cmd.Run()
		}()

		select {
		case <-timeout:
			_ = cmd.Process.Kill()
			task.Status = domain.StatusFailure
			logger.Errorf("task %s reached timeout limit", task.Name)
			return fmt.Errorf("task %s timeout", task.Name)
		case err := <-done:
			if err != nil {
				task.Status = domain.StatusFailure
				stdErrBuf.WriteString(err.Error())
				return err
			}
			task.Status = domain.StatusSuccess
			return nil
		}
	}

	// 无超时执行
	if err := cmd.Run(); err != nil {
		task.Status = domain.StatusFailure
		stdErrBuf.WriteString(err.Error())
		return err
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
	sErr := fmt.Sprintf("%s stderr: %s", task.Name, stdErr.String())
	sOut := fmt.Sprintf("%s stdout: %s", task.Name, stdOut.String())

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

	// 发送消息
	e.message.Send(sErr)
	e.message.Send(sOut)
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
