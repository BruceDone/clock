package controller

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"

	"clock/param"
)

type PieData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func GetLoadAverage(c echo.Context) error {
	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	v, e := load.Avg()
	if e != nil {
		resp.Msg = fmt.Sprintf("[get load average] error to get with: %v", e)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	piedata := make(map[string]int)
	piedata["load1"] = int(v.Load1)
	piedata["load5"] = int(v.Load5)
	piedata["load15"] = int(v.Load15)
	resp.Data = piedata

	return c.JSON(http.StatusOK, resp)
}

func GetMemoryUsage(c echo.Context) error {
	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	v, e := mem.VirtualMemory()
	if e != nil {
		resp.Msg = fmt.Sprintf("[get memory average] error to get with: %v", e)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	used := PieData{
		Name:  "已使用",
		Value: int(v.UsedPercent),
	}

	free := PieData{
		Name:  "未使用",
		Value: 100 - used.Value,
	}

	data := []PieData{used, free}
	resp.Data = data

	return c.JSON(http.StatusOK, resp)
}

func GetCpuUsage(c echo.Context) error {
	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	v, e := cpu.Percent(time.Duration(1*time.Second), true)
	if e != nil {
		resp.Msg = fmt.Sprintf("[get cpu usage] error to get the task param with: %v", e)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	result := make(map[string]int)
	for i := 0; i < len(v); i++ {
		name := fmt.Sprintf("CPU%v", i)
		value := int(v[i])
		result[name] = value
	}
	resp.Data = result
	return c.JSON(http.StatusOK, resp)

}
