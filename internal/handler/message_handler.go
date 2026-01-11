package handler

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"clock/internal/service"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	messageService service.MessageService
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(messageService service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// GetTaskStatus WebSocket推送任务状态
func (h *MessageHandler) GetTaskStatus(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	msgChan := h.messageService.Receive()

	for {
		select {
		case msg := <-msgChan:
			if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return nil
			}
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// GetMessages 获取任务统计
func (h *MessageHandler) GetMessages(c echo.Context) error {
	counters := h.messageService.GetCounters()
	return OK(c, counters)
}
