package domain

// 任务状态常量
const (
	StatusPending = iota + 1 // 等待中
	StatusStart              // 运行中
	StatusSuccess            // 成功
	StatusFailure            // 失败
)

// 数据库后端类型
const (
	DBBackendSQLite   = "sqlite3"
	DBBackendMySQL    = "mysql"
	DBBackendPostgres = "postgres"
)

// StatusText 返回状态文本
func StatusText(status int) string {
	switch status {
	case StatusPending:
		return "pending"
	case StatusStart:
		return "running"
	case StatusSuccess:
		return "success"
	case StatusFailure:
		return "failure"
	default:
		return "unknown"
	}
}
