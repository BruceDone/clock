package domain

// TaskLog 任务执行日志
type TaskLog struct {
	Lid      string `json:"lid" gorm:"primaryKey"`        // 日志ID
	Tid      int    `json:"tid" gorm:"index:idx_log_tid"` // 任务ID
	Cid      int    `json:"cid" gorm:"index:idx_log_cid"` // 容器ID
	StdOut   string `json:"std_out"`                      // 标准输出
	StdErr   string `json:"std_err"`                      // 标准错误
	UpdateAt int64  `json:"update_at" gorm:"index"`       // 创建时间
}

// TableName 指定表名
func (TaskLog) TableName() string {
	return "task_logs"
}
