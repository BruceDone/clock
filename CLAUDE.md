# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Clock 是一个基于 Go 的可视化任务调度框架，支持 DAG 任务依赖和 bash 命令执行。前后端打包后为单个二进制文件。

## Common Commands

```bash
# 开发运行
go run main.go -c config/dev.yaml

# 构建 (Linux/macOS/Windows)
make all

# 分别构建
make linux   # 输出 bin/clock-linux
make mac     # 输出 bin/clock-darwin
make win     # 输出 bin/clock-windows.exe

# 代码格式化
make fmt

# 测试
go test ./runner     # 仅 runner 测试 (storage 测试需要数据库环境)
```

## Architecture

### 分层结构
```
main.go → config → server → controller → storage/runner
```

- **config/** - YAML 配置加载 (Viper)
- **server/** - Echo 引擎初始化、路由、JWT 中间件
- **controller/** - HTTP API 处理器
- **storage/** - GORM 模型、cron 调度器、任务执行逻辑
- **runner/** - 任务执行器 (超时、中断支持)

### 核心模型 (storage/model.go)
- **Container** - 任务组，包含 cron 表达式
- **Task** - 单个 bash 命令任务
- **Relation** - DAG 依赖关系
- **TaskLog** - 任务输出日志

### 关键流程
1. **启动流程**: main.go 解析配置 → 初始化 DB → 启动 cron 调度器 → Echo 服务监听
2. **任务执行**: Container 触发 → `RunContainer()` → 拓扑排序 → 按依赖级别执行 `RunSingleTask()`
3. **API 分组**: `/v1/task`, `/v1/container`, `/v1/relation`, `/v1/log`, `/v1/node`, `/v1/message`, `/v1/system`

### WebSocket 实时更新
- 端点: `/v1/task/status`
- 通过 `storage.Messenger.Channel` 推送任务状态

## Database

默认使用 SQLite (`clock.db`)，支持 MySQL 和 PostgreSQL (GORM)。

## 注意事项

- 前端资源通过 `go:embed` 打包到二进制 (位于 `server/webapp/`)
- 默认登录: admin / admin (config/dev.yaml)
- 服务默认端口: 9528
