package storage

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/sirupsen/logrus"
)

func GenGuid(length int) (string, error) {
	u, e := uuid.NewV4()

	if e != nil {
		logrus.Error(e)
		return "", e
	}

	guid := u.String()
	guid = strings.Replace(guid, "-", "", -1)

	return guid[0:length], nil
}

// 根据tid 执行task ,
func RunSingleTaskByID(tid int) error {
	var task Task
	enable := true

	if err := Db.Where("tid = ?", tid).Find(&task).Error; err != nil {
		logrus.Errorf("[run single task by id] error to get task: %v", err)
		return err
	}

	// 判断上级任务是否执行成功
	var relations []Relation
	if err := Db.Where("cid = ?  and next_tid = ?", task.Cid, tid).Find(&relations).Error; err != nil {
		logrus.Errorf("[run single task by id] error to get relations: %v", err)
		return err
	}

	// 存在关系依赖，判断上级任务
	if len(relations) > 0 {
		var tids []int
		for _, item := range relations {
			tids = append(tids, item.Tid)
		}

		var tasks []Task
		if err := Db.Where("tid in (?)", tids).Find(&tasks).Error; err != nil {
			logrus.Errorf("[run single task by id] error to query tasks: %v", err)
			return err
		}

		// 存在失败情况
		for _, item := range tasks {
			if item.Status == FAILURE || item.Status == PENDING {
				enable = false
				break
			}
		}
	}

	// 执行任务
	if enable {
		return RunSingleTask(task)
	} else {
		task.Status = PENDING
		Db.Save(&task)
	}

	return nil
}

func RunSingleTask(t Task) error {
	var stdOutBuf bytes.Buffer
	var stdErrBuf bytes.Buffer

	// 保存最后状态
	t.Status = START
	defer func() {
		t.UpdateAt = time.Now().Unix()
		logrus.Debugf("[%v] - now finish task [%s]", t.Tid, t.Name)
		Db.Save(&t)
		saveLog(t, stdOutBuf, stdErrBuf)
	}()

	if t.Command == "" {
		t.Status = FAILURE
		return errors.New("please do not input the empty command")
	}

	logrus.Debugf("[%v] - now will run the task [%s]", t.Tid, t.Name)
	Db.Save(&t)

	args := strings.Split(t.Command, " ")
	c := exec.Command(args[0], args[1:]...)
	c.Stdout = &stdOutBuf
	c.Stderr = &stdErrBuf

	if t.Directory != "" {
		c.Dir = t.Directory
	}

	if t.TimeOut > 0 {
		timeout := time.After(time.Duration(t.TimeOut) * time.Second)
		done := make(chan error, 1)

		go func() {
			done <- c.Run()
		}()

		select {
		case <-timeout:
			_ = c.Process.Kill()
			logrus.Errorf("cmd %s reach to timeout limit", t.Command)
		case <-done:
			return nil
		}
	}

	e := c.Run()
	if e != nil {
		logrus.Error(e)
		// 写入错误信息
		stdErrBuf.WriteString(e.Error())
		t.Status = FAILURE
		return e
	}

	t.Status = SUCCESS
	return nil

}

// 如果一个节点是包含子节点
func RunContainer(c Container) error {
	// 保存状态
	c.Status = START
	Db.Save(&c)

	defer func() {
		c.Status = SUCCESS
		Db.Save(&c)
	}()

	var relations []Relation
	Db.Where(" cid = ?", c.Cid).Find(&relations)

	var tasks []Task
	Db.Where(" cid = ?", c.Cid).Find(&tasks)

	runStageTasks(tasks, relations)
	return nil
}

func runStageTasks(tasks []Task, relations []Relation) {
	stage := 0
	for {
		logrus.Debugf("[run container] now in stage %d", stage)
		// 终止条件
		if len(tasks) < 1 {
			break
		}
		var rootTids []int

		// 初使化入度
		query := make(map[int]int)
		for _, item := range tasks {
			query[item.Tid] = 0
		}

		// 计算入度
		for _, item := range relations {
			v, ok := query[item.NextTid]
			if !ok {
				continue
			}
			query[item.NextTid] = v + 1
		}

		// 筛选入度
		for k, v := range query {
			if v == 0 {
				rootTids = append(rootTids, k)
			}
		}

		// 存在环
		if len(rootTids) < 1 {
			logrus.Warn("[run container] exists the circle")
			break
		}

		for _, tid := range rootTids {
			if err := RunSingleTaskByID(tid); err != nil {
				logrus.Errorf("[run container] run task %d with error: %v", tid, err)
			}
		}

		// 移除节点
		var tmpTasks []Task
		for _, item := range tasks {
			if !contains(rootTids, item.Tid) {
				tmpTasks = append(tmpTasks, item)
			}
		}

		// 移除关系
		var tmpRelations []Relation
		for _, item := range relations {
			if !contains(rootTids, item.Tid) {
				tmpRelations = append(tmpRelations, item)
			}
		}

		tasks = tmpTasks
		relations = tmpRelations
		stage += 1
	}
}

func saveLog(t Task, stdOut, stdErr bytes.Buffer) {
	sErr := fmt.Sprintf("%s stderr is : %s", t.Name, stdErr.String())
	sOut := fmt.Sprintf("%s stdout is : %s", t.Name, stdOut.String())

	// 回写日志状态
	if t.LogEnable {
		lid, _ := GenGuid(8)
		l := TaskLog{
			Lid:      lid,
			Tid:      t.Tid,
			Cid:      t.Cid,
			StdOut:   stdOut.String(),
			StdErr:   stdErr.String(),
			UpdateAt: time.Now().Unix(),
		}
		Db.Save(&l)
	}

	sendMessage(sErr)
	sendMessage(sOut)
}

func sendMessage(msg string) {
	select {
	case Messenger.Channel <- msg:
	default:
		logrus.Warnf("the Messenger is full now ")
	}
}
