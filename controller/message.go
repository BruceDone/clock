package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"clock/param"
	"clock/storage"
)

func GetMessages(c echo.Context) error {
	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	counters, err := getMessages()
	if err != nil {
		resp.Msg = fmt.Sprintf("[get messages] error to get the counts: %v", err)
		return c.JSON(http.StatusBadRequest, resp)
	}
	resp.Data = counters

	return c.JSON(http.StatusOK, resp)
}

func getMessages() ([]storage.TaskCounter, error) {
	var tasks []storage.Task
	var counters []storage.TaskCounter

	if err := storage.Db.Find(&tasks).Error; err != nil {
		logrus.Errorf("[get messages] error to get tasks with error : %v", err)
		return counters, err
	}

	pending := storage.TaskCounter{
		Title: "当前等待",
		Icon:  "md-clock",
		Count: 0,
		Color: "#ff9900",
	}

	start := storage.TaskCounter{
		Title: "正在运行",
		Icon:  "md-play",
		Count: 0,
		Color: "#19be6b",
	}

	success := storage.TaskCounter{
		Title: "运行成功",
		Icon:  "md-done-all",
		Count: 0,
		Color: "#2d8cf0",
	}

	failure := storage.TaskCounter{
		Title: "运行失败",
		Icon:  "md-close",
		Count: 0,
		Color: "#ed3f14",
	}

	for _, t := range tasks {
		switch t.Status {
		case storage.PENDING:
			pending.Count += 1
		case storage.START:
			start.Count += 1
		case storage.SUCCESS:
			success.Count += 1
		case storage.FAILURE:
			failure.Count += 1
		default:
			logrus.Warnf("find the unknown tasks %v : status %v ", t.Tid, t.Status)
		}
	}

	counters = append(counters, pending, start, success, failure)
	return counters, nil
}
