package domain

// Message 消息通道
type Message struct {
	Size    int         // 容量
	Channel chan string // 信息通道
}

// NewMessage 创建消息通道
func NewMessage(size int) *Message {
	if size <= 0 {
		size = 1000
	}
	return &Message{
		Size:    size,
		Channel: make(chan string, size),
	}
}

// Send 发送消息（非阻塞）
func (m *Message) Send(msg string) {
	select {
	case m.Channel <- msg:
	default:
		// 通道满时丢弃消息
	}
}

// Receive 接收消息通道
func (m *Message) Receive() <-chan string {
	return m.Channel
}

// TaskCounter 任务统计
type TaskCounter struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
	Count int    `json:"count"`
	Color string `json:"color"`
}
