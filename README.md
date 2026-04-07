# OA System

基于 **Go + Vue 3** 的前后端分离 OA 办公系统，覆盖组织人事、公告、会议与日程、审批流程、菜单权限、操作日志等核心场景。

---

## 1. 项目结构

```text
oa-system/
├─ backend/                      # Go 后端（Gin + GORM + MySQL + Redis）
│  ├─ cmd/
│  ├─ config/                    # 配置加载
│  ├─ database/                  # MySQL / Redis 初始化
│  ├─ dto/                       # 请求/响应 DTO 定义
│  ├─ handlers/                  # 业务处理层
│  ├─ middleware/                # JWT 鉴权等中间件
│  ├─ models/                    # GORM 数据模型
│  ├─ routes/                    # 路由注册
│  ├─ uploads/                   # 上传文件目录
│  ├─ .env                       # 环境变量
│  ├─ go.mod
│  └─ main.go
├─ frontend/                     # Vue 前端（Vite + Element Plus）
│  ├─ src/
│  │  ├─ api/                    # 接口封装
│  │  ├─ router/                 # 前端路由
│  │  ├─ views/                  # 页面视图
│  │  ├─ styles/
│  │  └─ utils/
│  ├─ package.json
│  └─ vite.config.js
└─ README.md
```

---

## 2. 核心技术栈

### 后端
- Go 1.22+
- Gin（HTTP 框架）
- GORM + MySQL（ORM 与关系型数据库）
- Redis（Token 校验、缓存相关能力）
- JWT（登录鉴权）
- Orchid（流程引擎）
- Bluemonday（富文本安全清洗）

### 前端
- Vue 3
- Vite 5
- Vue Router 4
- Element Plus
- Axios
- WangEditor

---

## 3. 功能概览

- 用户登录、个人信息、修改密码
- 组织与角色管理：部门、岗位（角色）、部门-岗位关联、员工管理
- 公告管理（含审批流）
- 日程与会议：日历事件、会议室、事件预定
- 人事流程：入职、请假、离职（含提交、审批、撤回、取消审批）
- 工作流管理：流程模板、业务类型、Orchid 流程定义与任务处理
- 我的待办能力：待我审核、已审核、待阅、已阅
- 菜单与权限：动态菜单、岗位菜单权限
- 操作日志与异步日志写入
- 文件上传：图片与附件

---

## 4. 环境要求

- Go 1.22+
- Node.js `18+`
- npm `9+`
- MySQL `8+`
- Redis `6+`
---

## 5. 后端启动

### 5.1 配置环境变量
在 `backend/` 下创建并配置 `.env`：

```env
DB_HOST=localhost
DB_PORT=your_databaseport
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=oa_system
SERVER_PORT=your_serverport
JWT_SECRET=your_jwt_secret_key
REDIS_HOST=localhost
REDIS_PORT=your_redisport
REDIS_PASSWORD=your_redispassword
```

### 5.2 启动命令

```bash
cd backend
go mod tidy
go run main.go
```

后端默认监听：`http://localhost:your_serverport`

> 首次启动会自动迁移数据表，并初始化管理员账号（若库内无用户）：
> - 用户名：`admin`
> - 密码：`admin123`
> 
> 请首次登录后立即修改密码。

---

## 6. 前端启动

```bash
cd frontend
npm install
npm run dev
```

前端默认地址：`http://localhost:5173`

Vite 代理配置修改：
- `/api` -> `http://localhost:your_serverport`
- `/uploads` -> `http://localhost:your_serverport`

---

## 7. 构建

### 后端编译
```bash
cd backend
go build ./...
```

### 前端构建
```bash
cd frontend
npm run build
```

---

## 8. License

内部项目，按团队规范使用。