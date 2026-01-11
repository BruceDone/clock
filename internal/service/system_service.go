package service

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// systemService 系统监控服务实现
type systemService struct{}

// NewSystemService 创建系统监控服务
func NewSystemService() SystemService {
	return &systemService{}
}

// GetLoadAverage 获取系统负载
func (s *systemService) GetLoadAverage() ([]float64, error) {
	avg, err := load.Avg()
	if err != nil {
		return nil, err
	}
	return []float64{avg.Load1, avg.Load5, avg.Load15}, nil
}

// GetMemoryUsage 获取内存使用率
func (s *systemService) GetMemoryUsage() (float64, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return v.UsedPercent, nil
}

// GetCPUUsage 获取CPU使用率
func (s *systemService) GetCPUUsage() (float64, error) {
	percents, err := cpu.Percent(0, false)
	if err != nil {
		return 0, err
	}
	if len(percents) > 0 {
		return percents[0], nil
	}
	return 0, nil
}
