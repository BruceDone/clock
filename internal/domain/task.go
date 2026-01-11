package domain

// Task 任务实体
type Task struct {
	Tid       int    `json:"tid" gorm:"primaryKey"`         // 任务ID
	Cid       int    `json:"cid" gorm:"index:idx_task_cid"` // 容器ID
	Command   string `json:"command"`                       // bash命令
	Name      string `json:"name"`                          // 任务名称
	Directory string `json:"directory"`                     // 工作目录
	Disable   bool   `json:"disable"`                       // 是否禁用
	Status    int    `json:"status" gorm:"default:1"`       // 当前状态
	Timeout   int    `json:"timeout"`                       // 超时时间(秒)
	UpdateAt  int64  `json:"update_at"`                     // 修改时间
	LogEnable bool   `json:"log_enable"`                    // 是否启用日志
	PointX    int    `json:"point_x"`                       // 可视化坐标X
	PointY    int    `json:"point_y"`                       // 可视化坐标Y
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}

// IsDisabled 检查是否禁用
func (t *Task) IsDisabled() bool {
	return t.Disable
}

// SetStatus 设置状态
func (t *Task) SetStatus(status int) {
	t.Status = status
}

// Node 关系图节点（视图对象）
type Node struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

// ToNode 将Task转换为Node
func (t *Task) ToNode() Node {
	return Node{
		ID:     t.Tid,
		Name:   t.Name,
		Status: t.Status,
		X:      t.PointX,
		Y:      t.PointY,
	}
}
