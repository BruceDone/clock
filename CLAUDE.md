# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Clock 是一个基于 Go cron 的轻量级可视化调度框架，支持 DAG 任务依赖关系和 bash 命令执行。前后端通过 `go:embed` 打包成单个二进制文件。

## 常用命令

### 后端 (Go)
```bash
# 编译二进制（跨平台）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o clock cmd/clock/main.go

# 运行
./clock -c configs/config.toml

# 运行测试
go test ./...
```

### 前端 (Vue 3)
```bash
cd server/webapp

# 开发模式
npm run dev

# 构建生产版本（输出到 web/dist）
npm run build
```

## 架构

### 后端 (Go)
- **cmd/clock/main.go**: 程序入口；初始化数据库、StreamHub、服务、处理器，并启动调度器
- **internal/config/**: 从 TOML 配置文件加载配置（server、storage、auth、message）
- **internal/domain/**: 核心模型：Task、Container、Relation（DAG 边）、TaskLog、Message
- **internal/repository/**: 基于 GORM 的数据访问层
- **internal/service/**: 业务逻辑层
  - `SchedulerService`: 使用 robfig/cron 管理定时任务；启动/停止容器调度
  - `Executor`: 执行任务，通过拓扑排序处理 DAG 依赖
  - `StreamHub**: SSE 广播中心，实时推送任务日志到客户端
- **internal/handler/**: Echo HTTP 处理器，处理 API 端点
- **internal/router/**: 注册 API 路由（JWT 中间件保护）；提供嵌入式前端
- **internal/middleware/**: JWT 认证、请求日志

### 前端 (Vue 3 + TypeScript)
- **src/views/**: 页面组件（首页、容器列表、容器配置、任务列表、实时状态、日志中心）
- **src/api/**: Axios API 客户端
- **src/stores/**: Pinia 状态管理（用户认证状态）
- 使用 Element Plus UI 组件和 AntV G6 进行 DAG 图可视化

### 数据模型
- **Container**: 任务组，带 cron 表达式用于定时执行
- **Task**: 容器内的单个 bash 命令执行
- **Relation**: DAG 边，定义任务依赖关系（Tid -> NextTid）
- 任务按拓扑顺序执行；如果前置任务失败或等待中，当前任务设为等待状态

## 关键模式

### SSE 实时推送
`StreamHub` 管理订阅者频道；`Executor` 发布 `StreamEvent`（task_start、task_end、stdout、stderr）到所有订阅者。客户端通过 `/v1/message` SSE 端点连接。

### DAG 执行
`Executor.runStageTasks()` 使用 Kahn 算法实现拓扑排序。每一阶段并行执行所有入度为 0 的任务，然后移除已完成的节点及其出边。

### JWT 认证
Token 通过 Header（`Authorization: Bearer <token>`）、Cookie 或 Query 参数传递。在 `configs/config.toml` 的 `[auth]` 部分配置。

## 数据库支持
SQLite（默认）、MySQL、PostgreSQL。通过配置文件的 `[storage]` 部分设置。
