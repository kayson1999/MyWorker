# 👷 打工人打卡（MyWorker）

一款轻量级的员工打卡考勤应用，支持上下班打卡、工时统计、排行榜等功能。

## ✨ 功能特性

- 🕐 每日上下班打卡
- 📊 工时统计与分析
- 🏆 打卡排行榜
- 👤 个人资料管理
- 🔐 统一用户中心认证（MyUserCenter）

## 🛠️ 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Vue 3 + Vue Router + Vite |
| 后端 | Go (net/http) |
| 数据库 | SQLite (go-sqlite3) |
| 部署 | Docker / Docker Compose |

## 📁 项目结构

```
MyWorker/
├── src/                    # 前端源码
│   ├── components/         # Vue 组件
│   │   └── clockin/        # 打卡相关组件
│   ├── views/              # 页面视图
│   ├── services/           # API 服务 & 状态管理
│   ├── router/             # 路由配置
│   ├── styles/             # 全局样式
│   ├── config/             # 前端配置
│   ├── App.vue             # 根组件
│   └── main.js             # 入口文件
├── server/              # Go 后端源码
│   ├── main.go             # 服务入口
│   ├── config/             # 配置加载
│   ├── logger/             # 日志模块（按天滚动）
│   ├── db/                 # 数据库初始化
│   ├── routes/             # API 路由处理
│   ├── middleware/          # 中间件（认证等）
│   ├── usercenter/         # 用户中心 HTTP 客户端
│   ├── localuser/          # 本地业务用户管理
│   ├── go.mod              # Go 模块定义
│   └── .env                # 环境变量配置
├── public/                 # 静态资源
├── index.html              # HTML 入口
├── vite.config.js          # Vite 构建配置
├── Dockerfile              # Docker 镜像构建文件（多阶段）
├── docker-compose.yml      # Docker Compose 编排文件
├── start.sh                # 一键启动脚本
└── package.json            # 前端依赖配置
```

## 🚀 快速开始

### 环境要求

- Node.js >= 18（前端构建）
- Go >= 1.21（后端）
- GCC（编译 go-sqlite3 需要 CGO）
- （可选）Docker & Docker Compose

### 本地开发

1. **安装依赖**

```bash
# 安装前端依赖
npm install

# 下载 Go 模块
cd server && go mod tidy && cd ..
```

2. **配置环境变量**

编辑 `server/.env` 文件：

```env
PORT=9009
USER_CENTER_URL=http://localhost:4000
USER_CENTER_APP_ID=myworker

# 日志配置
LOG_LEVEL=info
LOG_MAX_FILES=1
LOG_TO_CONSOLE=true
```

3. **启动开发服务**

```bash
# 使用启动脚本（推荐）
./start.sh start

# 或手动启动：
# 终端 1 — 启动 Go 后端
cd server && go run .

# 终端 2 — 启动前端开发服务器
npm run dev
```

- 前端开发服务器：`http://localhost:3002`
- 后端 API 服务：`http://localhost:9009`
- 前端开发模式下 `/api` 请求会自动代理到后端

### 生产构建

```bash
# 使用启动脚本
./start.sh build

# 或手动构建：
npm run build
cd server && CGO_ENABLED=1 go build -o ../myworker-server .

# 启动
./myworker-server
```

访问 `http://localhost:9009` 即可使用。

## 🐳 Docker 部署

### 方式一：Docker Compose（推荐）

```bash
# 一键构建并启动
docker compose up -d

# 查看日志
docker compose logs -f

# 停止服务
docker compose down
```

### 方式二：手动构建镜像

```bash
# 构建镜像
docker build -t myworker .

# 运行容器
docker run -d \
  --name myworker \
  -p 9009:9009 \
  -v myworker-data:/app/db \
  -v myworker-logs:/app/logs \
  -e USER_CENTER_URL=http://host.docker.internal:4000 \
  -e USER_CENTER_APP_ID=myworker \
  myworker
```

部署完成后访问 `http://localhost:9009` 即可使用。

## ⚙️ 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `PORT` | 服务监听端口 | `9009` |
| `USER_CENTER_URL` | 用户中心服务地址 | `http://localhost:4000` |
| `USER_CENTER_APP_ID` | 本服务在用户中心的租户标识 | `myworker` |
| `LOG_DIR` | 日志目录 | 启动根路径/logs |
| `LOG_LEVEL` | 日志级别：debug / info / warn / error | `info` |
| `LOG_MAX_FILES` | 最多保留的日志文件数量（按天滚动） | `1` |
| `LOG_TO_CONSOLE` | 是否同时输出到控制台 | `true` |

## 📋 日志设计

- **滚动策略**：按天滚动，每天一个日志文件，文件名格式 `app-YYYY-MM-DD.log`
- **文件数量**：通过 `LOG_MAX_FILES` 配置，默认最多保留 1 个日志文件
- **日志目录**：默认为启动根路径下的 `logs/` 目录，可通过 `LOG_DIR` 配置
- **日志格式**：`[时间] [级别] [文件名:行号@函数名] 消息内容`
- **示例输出**：
  ```
  [2026-03-26 20:30:00.123] [INFO ] [main.go:45@main] 🚀 打工人打卡服务已启动: http://localhost:9009/worker
  [2026-03-26 20:30:01.456] [ERROR] [auth.go:35@handleLogin] 登录失败: 用户名或密码错误
  ```

## 📝 API 接口

| 路径 | 说明 |
|------|------|
| `GET /api/health` | 健康检查 |
| `/api/auth/*` | 认证相关（注册、登录、登出） |
| `/api/user/*` | 用户信息（获取、更新） |
| `/api/clockin/*` | 打卡操作（上班、下班、补卡、调整、统计） |
| `/api/ranking/*` | 排行榜（工时、早起、夜猫、连续打卡、准时） |

## 📌 注意事项

- 本服务依赖 **MyUserCenter** 统一用户中心进行认证，部署前请确保用户中心服务已启动
- SQLite 数据库文件存储在 `db/clockin.db`，Docker 部署时通过 Volume 持久化
- Docker 容器内访问宿主机服务请使用 `host.docker.internal` 地址
- Go 后端编译需要 CGO 支持（因为 go-sqlite3 是 C 绑定），确保系统安装了 GCC
