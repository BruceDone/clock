<div align=center>
<img src="https://user-images.githubusercontent.com/12979090/86565300-297abd80-bf9a-11ea-916f-b547f5023ee8.png" /> 
</div>

## Clock
基于go cron的可视化调度轻量级调度框架，支持DAG任务依赖，支持bash命令，前端及后端编译完成(基于packr2)后仅有一个二进制文件，轻松部署

## 地址
* 后台: https://github.com/BruceDone/clock
* 前台: https://github.com/BruceDone/clock-admin

## 环境
* 后端
    * go 1.20
    * [packr](https://github.com/gobuffalo/packr) - 静态文件打包
    * [cron](https://github.com/robfig/cron) - 定时调度器
    * [echo](https://github.com/labstack/echo)
    * [gorm](https://github.com/jinzhu/gorm)
* 前端
    * vue 
    * [iview-admin](https://github.com/iview/iview-admin)

## 使用
### 直接使用
下载git上的release列表，根据系统下载相应的二进制文件，使用命令
```
./clock -c ./config/dev.yaml
```

### 自己编译前后端
将前端项目 clock-admin 下载到本地，使用命令 `npm run build`, 编译生成前端项目文件`dist`, 将后端项目 clock 下载到本地, 进入项目根目录，确保安装了packr2 ,使用如下命令

```shell script
rm -rf webapp
mkdir -p ./webapp
cp -r /你的clock-admin文件夹/dist/* ./webapp
packr2 clean
packr2
# 根据发布的目标平台，调整如下命令
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go generate  
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
```

使用命令`./clock -c config/dev.yaml` 载入你的配置文件，完成后打开浏览器 `http://127.0.0.1:9528` ，输入用户名密码就可以进入管理后台
![login](https://user-images.githubusercontent.com/12979090/86568293-3948d080-bf9f-11ea-9c19-4cf68af595a0.png)

## 结构
```
├── config - 配置文件夹,示例文件所在
├── controller - 控制层
├── packrd - packr2生成的静态资源
├── param - 参数相关
├── runner - 执行器
├── server - view层
├── storage - 存储相关
└── webapp - 由clock-admin发布的前端资源
    ├── css
    ├── fonts
    ├── img
    └── js
```

## 特性与功能
* [DAG任务关联](https://en.wikipedia.org/wiki/Dag) , 可以管理任务的前后依赖
* 可视化管理
* 支持多种数据库: sqlite , mysql ,postgresql
* 前后端打包完成后只有一个二进制文件，极其方便部署
* 跨平台

## 使用截图

### 登录进入控制台 
![personal](https://user-images.githubusercontent.com/12979090/86567691-5c26b500-bf9e-11ea-8c3c-98a75120ce18.jpg)

### 添加任务容器
![fathertask](https://user-images.githubusercontent.com/12979090/86567720-6779e080-bf9e-11ea-9168-18dc751d730e.jpg)

点击新增，调度表达式这里支持cron和@every语法，更多语法请参考:[cron](https://github.com/robfig/cron)

### 点击配置进入子任务配置界面
![taskdag](https://user-images.githubusercontent.com/12979090/86567779-7a8cb080-bf9e-11ea-8622-fc924f4a5ba8.jpg)

点击任务编辑下的新增，选中新增的节点，编辑任务bash命令，任务名，是否保存日志，及任务超时时间(小技巧:选中画板空白处为新增，选中节点为编辑状态)，可以自由编辑节点(任务)之间的关系，摆好位置之后选择保存

### 查看后台任务输出日志
![status](https://user-images.githubusercontent.com/12979090/86567810-84aeaf00-bf9e-11ea-82b6-4bd585d7df7c.jpg)

### 查看持久化的日志
![loglist](https://user-images.githubusercontent.com/12979090/86567837-8e381700-bf9e-11ea-9812-43a7189a2827.jpg)