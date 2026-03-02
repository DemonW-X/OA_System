# OA System

一个基于 **Go + Vue3** 的前后端分离 OA 办公系统，包含组织人事、公告、日程会议、审批流程、菜单管理、操作日志等功能。

---

## 1. 项目结构

```text
oa-system/
├─ backend/                 # Go 后端（Gin + GORM + MySQL + Redis）
│  ├─ cmd/
│  ├─ config/
│  ├─ database/
│  ├─ handlers/
│  ├─ middleware/
│  ├─ models/
│  ├─ routes/
│  ├─ uploads/
│  ├─ .env
│  └─ main.go
└─ frontend/                # Vue3 前端（Vite + Element Plus）
   ├─ src/
   ├─ package.json
   └─ vite.config.js
```

---

## 2. 核心技术栈

### 后端
- Go 1.22+
- Gin
- GORM
- MySQL
- Redis
- JWT（登录鉴权）
- Orchid（流程引擎）

### 前端
- Vue 3
- Vite
- Vue Router
- Element Plus
- Axios
- WangEditor

---

## 3. 功能概览

- 用户登录、个人信息、修改密码
- 部门/职位/员工管理
- 公告管理
- 日历与会议室、事件预定
- 请假、离职管理
- 流程定义与审批（含待办）
- 菜单管理（动态菜单）
- 操作日志
- 文件上传（图片/附件）

---

## 4. 环境要求

- Go 1.22+
- Node.js 18+
- MySQL 8+
- Redis 6+

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

后端默认监听：`http://localhost:8080`

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

Vite 代理配置：
- `/api` -> `http://localhost:8080`
- `/uploads` -> `http://localhost:8080`

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