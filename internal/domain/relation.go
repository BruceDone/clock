package domain

// Relation 任务关系实体（DAG边）
type Relation struct {
	Rid      int   `json:"rid" gorm:"primaryKey"`                     // 关系ID
	Cid      int   `json:"cid" gorm:"index:idx_relation_cid"`         // 容器ID
	Tid      int   `json:"tid" gorm:"uniqueIndex:uidx_tid_next"`      // 前置任务ID
	NextTid  int   `json:"next_tid" gorm:"uniqueIndex:uidx_tid_next"` // 后续任务ID
	UpdateAt int64 `json:"update_at"`                                 // 修改时间
}

// TableName 指定表名
func (Relation) TableName() string {
	return "relations"
}

// Link 关系图边（视图对象）
type Link struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Cid     int    `json:"cid"`
	Tid     int    `json:"tid"`
	NextTid int    `json:"next_tid"`
}

// ToLink 将Relation转换为Link
func (r *Relation) ToLink() Link {
	return Link{
		ID:      r.Rid,
		Cid:     r.Cid,
		Tid:     r.Tid,
		NextTid: r.NextTid,
	}
}

// RelationGraph 关系图（视图对象）
type RelationGraph struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}
