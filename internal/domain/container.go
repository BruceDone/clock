package domain

// Container 任务容器实体
type Container struct {
	Cid        int    `json:"cid" gorm:"primaryKey"`        // 主键
	EntryID    int    `json:"entry_id"`                     // cron生成的调度ID
	Name       string `json:"name"`                         // 名称
	Expression string `json:"expression"`                   // cron表达式
	Status     int    `json:"status" gorm:"default:1"`      // 当前状态
	Disable    bool   `json:"disable"`                      // 是否禁用
	Blocking   bool   `json:"blocking" gorm:"default:true"` // 阻塞模式（上次未完成则跳过）
	UpdateAt   int64  `json:"update_at"`                    // 修改时间
}

// TableName 指定表名
func (Container) TableName() string {
	return "containers"
}
