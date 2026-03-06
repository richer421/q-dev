# 前端脚手架设计

## 定位

B 端平台，基于 Ant Design Pro 官方脚手架。

## 技术选型

- **框架**: Ant Design Pro（umi max）
- **UI**: antd 5.x 最新稳定版（脚手架自带）
- **包管理器**: pnpm
- **目录**: `frontend/`，与 `backend/` 平级

## 初始化策略

使用 `pro create` 生成完整脚手架后，清理 demo 内容：

- 移除示例页面（Welcome、TableList、Admin 等）
- 移除 mock 数据
- 保留框架骨架：ProLayout、路由配置、请求层

## 后端适配

后端统一响应格式：

```json
{"code": 0, "message": "ok", "data": ...}
```

配置 umi request 拦截器：
- `code === 0` 视为成功，返回 `data`
- `code !== 0` 视为业务错误，抛出 `message`
- HTTP 非 200 视为网络错误

## 后端 API 地址

开发环境通过 `proxy` 代理到 `http://localhost:8080`（后端 Gin 默认端口）。
