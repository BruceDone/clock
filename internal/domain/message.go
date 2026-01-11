package domain

// TaskCounter 任务统计
type TaskCounter struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
	Count int    `json:"count"`
	Color string `json:"color"`
}
