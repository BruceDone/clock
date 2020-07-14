package storage

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"clock/config"
	"clock/param"
)

var (
	Db          *gorm.DB
	scheduler   *cron.Cron
	Messenger   Message
	MessageSize = 1000
)

const (
	PENDING = iota + 1
	START
	SUCCESS
	FAILURE
)

type (
	// 容器ID
	Container struct {
		Cid        int    `json:"cid" gorm:"PRIMARY_KEY"`  // 主键
		EntryId    int    `json:"entry_id"`                // 由cron 生成的id
		Name       string `json:"name"`                    // 名字
		Expression string `json:"expression"`              // 表达式 支持@every [1s | 1m | 1h ] 参考 cron
		Status     int    `json:"status" gorm:"default:1"` // 当前状态
		Disable    bool   `json:"disable"`                 // 禁用
		UpdateAt   int64  `json:"update_at"`               // 修改时间
	}

	// 当前任务
	Task struct {
		Tid       int    `json:"tid" gorm:"PRIMARY_KEY"`        // task id
		Cid       int    `json:"cid" gorm:"index:task_idx_cid"` // 容器id
		Command   string `json:"command"`                       // 当前只支持bash command
		Name      string `json:"name"`                          // task 名字
		Directory string `json:"directory"`                     // 命令所在的目录
		Disable   bool   `json:"disable"`                       // 是否禁用当前任务
		Status    int    `json:"status" gorm:"default:1"`       // 当前状态
		TimeOut   int    `json:"timeout"`                       // 超时时间
		UpdateAt  int64  `json:"update_at"`                     // 修改时间
		LogEnable bool   `json:"log_enable"`                    // 是否启用日志
		PointX    int    `json:"point_x"`                       // 坐标 x 信息
		PointY    int    `json:"point_y"`                       // 坐标 y 信息
	}

	// 任务关系
	Relation struct {
		Rid      int   `json:"rid" gorm:"PRIMARY_KEY"`                    // 关系ID
		Cid      int   `json:"cid" gorm:"index"`                          // 容器ID
		Tid      int   `json:"tid" gorm:"unique_index:uidx_tid_sid"`      // 任务ID
		NextTid  int   `json:"next_tid" gorm:"unique_index:uidx_tid_sid"` // 子任务ID
		UpdateAt int64 `json:"update_at"`                                 // 修改时间
	}

	// 任务日志
	TaskLog struct {
		Lid      string `json:"lid"  gorm:"PRIMARY_KEY"`  // 主键Key
		Tid      int    `json:"tid" gorm:"index:idx_tid"` // task id
		Cid      int    `json:"cid" gorm:"index"`         // 绑定容器ID
		StdOut   string `json:"std_out"`                  // 正常输出流
		StdErr   string `json:"std_err"`                  // 异常输出流
		UpdateAt int64  `json:"update_at" gorm:"index"`   // 创建时间
	}
)

type (
	Page struct {
		Count   int    `json:"count"`
		Index   int    `json:"index"`
		Total   int    `json:"total"`
		Order   string `json:"order"`
		LeftTs  int64  `json:"left_ts" query:"left_ts"`
		RightTs int64  `json:"right_ts" query:"right_ts"`
	}

	// task 查询参数
	TaskQuery struct {
		Page
		Task
	}

	ContainerQuery struct {
		Page
		Container
	}

	LogQuery struct {
		Page
		TaskLog
	}
)

// 应用所需实体
type (
	// 关系Node节点
	Node struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Status int    `json:"status"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
	}

	// 关系Link
	Link struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Cid     int    `json:"cid"`      // 容器ID
		Tid     int    `json:"tid"`      // 当前节点ID
		NextTid int    `json:"next_tid"` // 子任务ID
	}

	Message struct {
		Size    int         //容量
		Channel chan string //信息通道
	}

	// 统计数据
	TaskCounter struct {
		Title string `json:"title"`
		Icon  string `json:"icon"`
		Count int    `json:"count"`
		Color string `json:"color"`
	}
)

// 初使化
func SetDb() {
	conn := config.Config.GetString("storage.conn")
	if conn == "" {
		logrus.Fatal("empty conn string")
	}

	backend := config.Config.GetString("storage.backend")
	if backend == "" {
		logrus.Fatal("not find the backend type")
	}

	var err error
	Db, err = gorm.Open(backend, conn)

	if err != nil {
		logrus.Fatal(err)
	}

	Db.AutoMigrate(&Task{}, &Container{}, &TaskLog{}, &Relation{})
	newScheduler()

	tmp := config.Config.GetInt("message.size")
	if tmp > 0 {
		MessageSize = tmp
	}
	Messenger = NewMessenger(MessageSize)
}

// 初使化信息通道
func NewMessenger(size int) Message {
	return Message{
		Size:    size,
		Channel: make(chan string, size),
	}
}

func newScheduler() {
	optLogs := cron.WithLogger(
		cron.VerbosePrintfLogger(
			log.New(os.Stdout, "[Cron]: ", log.LstdFlags)))

	scheduler = cron.New(optLogs)
	var containers []Container
	Db.Find(&containers)

	if len(containers) > 0 {
		for _, c := range containers {
			// 默认清空之前的状态
			c.Status = PENDING
			c.EntryId = -1
			c.UpdateAt = time.Now().Unix()
			if err := PutContainer(c); err != nil {
				logrus.Fatalf("[scheduler] error to init the task with error %v", err)
			}
		}
	}

	logrus.Info("[scheduler] start the ticker")
	scheduler.Start()
}

func GetWhereDb(object interface{}, filter []string) *gorm.DB {
	db := Db
	// 过滤bool 和子类型为struct内容
	filterKind := []string{"bool", "struct"}
	// 过滤Page参数体
	filterStruct := []string{"Page"}

	s := structs.New(object)
	for _, key := range s.Names() {
		tmp := s.Field(key)

		if inCondition(filterStruct, key) {
			continue
		}

		fields := tmp.Fields()
		for _, f := range fields {
			field := f.Tag("json")

			// 过滤的字段
			if inCondition(filter, field) {
				continue
			}

			kind := fmt.Sprintf("%v", f.Kind())
			// 过滤bool类型和struct类型
			if inCondition(filterKind, kind) {
				continue
			}

			value := fmt.Sprintf("%v", f.Value())
			if kind == "string" && value != "" {
				db = db.Where(fmt.Sprintf("%v like ?", field), "%"+value+"%")
			}

			if kind == "int" && value != "0" {
				db = db.Where(fmt.Sprintf("%v = ?", field), value)
			}

		}
	}

	return db
}

// 默认取出根任务
func GetTasks(query *TaskQuery) ([]Task, error) {
	var tasks []Task

	if query.Count < 1 {
		query.Count = 10
	}

	if query.Index < 1 {
		query.Index = 1
	}

	queryDB := GetWhereDb(query, nil)
	if e := queryDB.Model(tasks).Count(&query.Total).Error; e != nil {
		logrus.Error("failed to get the page total of tasks :" + e.Error())
		return nil, e
	}

	queryDB = queryDB.Offset((query.Index - 1) * query.Count).Limit(query.Count)

	if query.Order != "" {
		queryDB = queryDB.Order(query.Order)
	}

	if err := queryDB.Find(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetContainers(query *ContainerQuery) ([]Container, error) {
	var container []Container

	if query.Count < 1 {
		query.Count = 10
	}

	if query.Index < 1 {
		query.Index = 1
	}

	queryDB := GetWhereDb(query, nil)
	if e := queryDB.Model(container).Count(&query.Total).Error; e != nil {
		logrus.Error("failed to get the page total of containers :" + e.Error())
		return nil, e
	}

	queryDB = queryDB.Offset((query.Index - 1) * query.Count).Limit(query.Count)

	if query.Order != "" {
		queryDB = queryDB.Order(query.Order)
	}

	if err := queryDB.Find(&container).Error; err != nil {
		return nil, err
	}

	return container, nil
}

// 更新query 多页的情况
func GetLogs(query *LogQuery) ([]TaskLog, error) {
	var logs []TaskLog

	if query.Count < 1 {
		query.Count = 10
	}

	if query.Index < 1 {
		query.Index = 1
	}

	queryDB := GetWhereDb(query, []string{"lid"})
	if query.LeftTs > 0 {
		queryDB = queryDB.Where("update_at > ?", query.LeftTs)
	}

	if query.RightTs > 0 {
		queryDB = queryDB.Where("update_at < ?", query.RightTs)
	}

	if e := queryDB.Model(logs).Count(&query.Total).Error; e != nil {
		logrus.Error("failed to get the page total of logs :" + e.Error())
		return nil, e
	}

	queryDB = queryDB.Offset((query.Index - 1) * query.Count).Limit(query.Count).Order("update_at desc")
	if err := queryDB.Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// 根据ts 删除多久以前的数据
func DeleteLogs(query *LogQuery) error {
	queryDB := GetWhereDb(query, []string{"lid"})

	if query.LeftTs > 0 {
		queryDB = queryDB.Where("update_at > ?", query.LeftTs)
	}

	if query.RightTs > 0 {
		queryDB = queryDB.Where("update_at < ?", query.RightTs)
	}

	queryDB.LogMode(true)
	return queryDB.Delete(TaskLog{}).Error
}

// 返回相关内容
func GetRelations(cid int) (param.RelationResponse, error) {
	var relations []Relation
	var tasks []Task
	resp := param.RelationResponse{}

	err := Db.Where("cid = ?", cid).Find(&tasks).Error
	if err != nil {
		return resp, err
	}

	if len(tasks) < 1 {
		return resp, nil
	}

	if err := Db.Where(" cid = ?", cid).Find(&relations).Error; err != nil {
		return resp, nil
	}

	return makeRelation(tasks, relations)

}

// 检测容器内是否存在闭环
func CheckCircle(tasks []Task, relations []Relation) bool {
	var result = false

	if len(tasks) < 1 || len(relations) < 1 {
		return result
	}

	for {
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
			result = true
			break
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
	}

	return result
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func inCondition(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// 更新坐标信息
func PutNodes(nodes []Node) error {
	for _, v := range nodes {
		t, e := GetTask(v.ID)
		if e != nil {
			continue
		}

		t.PointX = v.X
		t.PointY = v.Y

		Db.Save(&t)
	}
	return nil
}

func makeRelation(tasks []Task, relations []Relation) (param.RelationResponse, error) {
	var nodes []Node
	var links []Link

	resp := param.RelationResponse{}

	if len(tasks) < 1 {
		return resp, nil
	}

	for _, item := range tasks {
		node := Node{
			ID:     item.Tid,
			Name:   item.Name,
			Status: item.Status,
			X:      item.PointX,
			Y:      item.PointY,
		}
		nodes = append(nodes, node)
	}

	for _, item := range relations {
		link := Link{
			ID:      item.Rid,
			Name:    "",
			Tid:     item.Tid,
			NextTid: item.NextTid,
		}
		links = append(links, link)
	}

	resp.Nodes = nodes
	resp.Links = links

	return resp, nil
}

func GetTask(tid int) (Task, error) {
	var t Task

	if err := Db.Where("tid = ?", tid).First(&t).Error; err != nil {
		return t, err
	}
	return t, nil
}

func GetContainer(cid int) (Container, error) {
	var c Container

	if err := Db.Where("cid = ?", cid).First(&c).Error; err != nil {
		return c, err
	}

	return c, nil
}

// 直接执行任务
func RunTask(tid int) error {
	task, err := GetTask(tid)
	if err != nil {
		logrus.Errorf("error to find the task with: %v", err)
		return err
	}

	return RunSingleTask(task)
}

// 删除容器 - 删除关联task
func DeleteContainer(cid int) error {
	c, e := GetContainer(cid)

	if e != nil {
		logrus.Errorf("[delete container] error to find the container %v with error : %v", cid, e)
		return e
	}

	scheduler.Remove(cron.EntryID(c.EntryId))

	if e := Db.Delete(c).Error; e != nil {
		logrus.Errorf("[delete container] error to delete the container %v with error : %v", cid, e)
		return e
	}

	if e := Db.Delete(Task{}, " cid = ?", cid).Error; e != nil {
		logrus.Errorf("[delete container] error to find the tasks with cid %v with error : %v", cid, e)
		return e
	}

	return nil
}

// 删除任务
func DeleteTask(tid int) error {
	// 使用事物进行原子操作
	return Db.Transaction(func(tx *gorm.DB) error {
		if err := Db.Where(" tid = ?", tid).Delete(Task{}).Error; err != nil {
			return err
		}

		if err := Db.Where(" tid = ?", tid).Delete(Relation{}).Error; err != nil {
			return err
		}

		if err := Db.Where(" next_tid = ?", tid).Delete(Relation{}).Error; err != nil {
			return err
		}
		// 回填用户token状态
		return nil
	})
}

// 更新任务或者新增任务
func PutTask(t *Task) error {
	t.UpdateAt = time.Now().Unix()
	return Db.Save(&t).Error
}

// 新增关系
func AddRelation(r *Relation) error {
	return Db.Save(&r).Error
}

// 移除状态
func DeleteRelation(rid int) error {
	return Db.Where("rid = ?", rid).Delete(&Relation{}).Error
}

// 添加任务，需要传入指针,方便修改值
func AddScheduler(c *Container) error {
	f := func() {
		if e := RunContainer(*c); e != nil {
			logrus.Error(e)
		}
	}

	entryId, e := scheduler.AddFunc(c.Expression, f)
	if e != nil {
		logrus.Error(e)
		return e
	}

	c.EntryId = int(entryId)
	logrus.Infof("[add scheduler] add the job of %s , with entry id %v", c.Name, c.EntryId)

	return nil
}

func PutContainer(c Container) error {
	// 移除并重新启用
	scheduler.Remove(cron.EntryID(c.EntryId))
	c.UpdateAt = time.Now().Unix()
	if c.Disable {
		c.EntryId = -1
	} else {
		err := AddScheduler(&c)
		if err != nil {
			logrus.Errorf("[put task] error with %v", err)
		}
	}

	return Db.Save(&c).Error
}
