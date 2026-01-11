package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

// GetTaskStatus SSE 推送任务状态（结构化事件流）
func (h *MessageHandler) GetTaskStatus(c echo.Context) error {
	res := c.Response()
	res.Header().Set(echo.HeaderContentType, "text/event-stream")
	res.Header().Set(echo.HeaderCacheControl, "no-cache")
	res.Header().Set("Connection", "keep-alive")
	res.Header().Set("X-Accel-Buffering", "no")
	res.WriteHeader(http.StatusOK)

	flusher, ok := res.Writer.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "streaming unsupported")
	}

	ctx := c.Request().Context()
	events := h.messageService.Subscribe(ctx)

	// initial comment to ensure the stream is established
	_, _ = fmt.Fprint(res.Writer, ": ok\n\n")
	flusher.Flush()

	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-keepalive.C:
			_, _ = fmt.Fprint(res.Writer, ": ping\n\n")
			flusher.Flush()
		case ev, ok := <-events:
			if !ok {
				return nil
			}

			data, err := json.Marshal(ev)
			if err != nil {
				continue
			}

			_, _ = fmt.Fprintf(res.Writer, "id: %d\n", ev.ID)
			_, _ = fmt.Fprint(res.Writer, "event: log\n")
			_, _ = fmt.Fprintf(res.Writer, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// GetMessages 获取任务统计
func (h *MessageHandler) GetMessages(c echo.Context) error {
	counters := h.messageService.GetCounters()
	return OK(c, counters)
}
