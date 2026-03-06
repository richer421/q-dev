# 前端脚手架实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在 `frontend/` 目录下搭建 Ant Design Pro 脚手架，清理 demo 内容，适配后端 API 格式。

**Architecture:** 使用 `pnpm create umi` 选择 Ant Design Pro 模板生成项目，然后清理示例页面和 mock，配置 request 拦截器对接后端 `{code, message, data}` 统一响应格式。

**Tech Stack:** Ant Design Pro / umi max / antd 5.x / pnpm

**Design Doc:** `docs/plans/2026-03-06-frontend-scaffold-design.md`

---

### Task 1: 安装 pnpm

当前环境只有 npm（Node v24.13.0），需先安装 pnpm。

**Step 1: 安装 pnpm**

```bash
npm install -g pnpm
```

**Step 2: 验证安装**

Run: `pnpm --version`
Expected: 版本号输出（如 9.x）

---

### Task 2: 生成 Ant Design Pro 脚手架

**Step 1: 进入 frontend 目录，执行脚手架生成**

```bash
cd /Users/richer/richer/q-dev/frontend
pnpm create umi
```

交互选项依次选择：
1. Pick Umi App Type → **Ant Design Pro**
2. Pick Npm Client → **pnpm**
3. Pick Npm Registry → **taobao**（国内环境）

**Step 2: 验证项目生成成功**

Run: `ls frontend/package.json`
Expected: 文件存在

**Step 3: 验证依赖安装和 dev server 能启动**

Run: `cd frontend && pnpm dev`
Expected: 编译成功，终端显示 `http://localhost:8000`（Ctrl+C 退出）

**Step 4: Commit**

```bash
git add frontend/
git commit -m "feat(frontend): scaffold Ant Design Pro with umi max"
```

---

### Task 3: 清理 demo 页面和 mock 数据

生成的脚手架包含大量示例内容，需要清理到只剩骨架。

**Step 1: 删除 demo 页面**

删除以下目录（如果存在）：

```bash
cd frontend
rm -rf src/pages/Welcome
rm -rf src/pages/TableList
rm -rf src/pages/Admin
rm -rf src/pages/User/Login
rm -rf src/pages/User/Register
rm -rf src/pages/exception
rm -rf src/pages/account
rm -rf src/pages/dashboard
rm -rf src/pages/form
rm -rf src/pages/list
rm -rf src/pages/profile
rm -rf src/pages/result
rm -rf src/pages/editor
```

注意：保留 `src/pages/` 目录本身和 `src/pages/404.tsx`（如果有）。

**Step 2: 清理 mock 数据**

```bash
rm -rf mock/*
```

**Step 3: 创建占位首页**

创建 `frontend/src/pages/Home/index.tsx`：

```tsx
const Home: React.FC = () => {
  return <div>Q-DEV</div>;
};

export default Home;
```

**Step 4: 简化路由配置**

修改 `frontend/config/routes.ts`，替换为：

```ts
export default [
  {
    path: '/',
    redirect: '/home',
  },
  {
    name: '首页',
    path: '/home',
    component: './Home',
  },
  {
    path: '*',
    layout: false,
    component: './404',
  },
];
```

注意：如果 404 页面不存在，创建一个简单的 `src/pages/404.tsx`：

```tsx
import { Result, Button } from 'antd';
import { history } from '@umijs/max';

const NotFound: React.FC = () => (
  <Result
    status="404"
    title="404"
    subTitle="页面不存在"
    extra={<Button type="primary" onClick={() => history.push('/')}>返回首页</Button>}
  />
);

export default NotFound;
```

**Step 5: 验证编译通过**

Run: `cd frontend && pnpm dev`
Expected: 编译成功，能看到 Q-DEV 首页

**Step 6: Commit**

```bash
git add -A
git commit -m "chore(frontend): remove demo pages and mock data"
```

---

### Task 4: 配置 request 拦截器适配后端 API

**Files:**
- Modify: `frontend/src/requestErrorConfig.ts`（或 `src/app.tsx` 中的 request 配置）

**Step 1: 查看当前 request 配置位置**

查看 `src/app.tsx` 和 `src/requestErrorConfig.ts`，确认 request 配置的位置。

**Step 2: 修改 request 配置**

在 request 配置中，设置响应拦截器匹配后端格式：

```ts
// 后端统一响应格式: { code: number, message: string, data?: any }
// code === 0 为成功，其他为失败
export const request: RequestConfig = {
  baseURL: '/api',
  errorConfig: {
    errorThrower: (res: any) => {
      const { code, message } = res;
      if (code !== 0) {
        const error: any = new Error(message);
        error.name = 'BizError';
        error.info = res;
        throw error;
      }
    },
    errorHandler: (error: any) => {
      if (error.name === 'BizError') {
        const errorInfo = error.info;
        notification.error({
          message: '请求失败',
          description: errorInfo.message,
        });
      } else if (error.response) {
        notification.error({
          message: `HTTP ${error.response.status}`,
          description: '网络请求异常',
        });
      } else {
        notification.error({
          message: '网络异常',
          description: '网络连接失败，请检查网络',
        });
      }
    },
  },
};
```

**Step 3: 配置开发代理**

修改 `frontend/config/proxy.ts`，配置 dev 环境代理：

```ts
export default {
  dev: {
    '/api/': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
};
```

**Step 4: 验证编译通过**

Run: `cd frontend && pnpm dev`
Expected: 编译成功

**Step 5: Commit**

```bash
git add -A
git commit -m "feat(frontend): configure request interceptor and API proxy"
```

---

### Task 5: 更新 .gitignore 和 knowledge 文档

**Step 1: 更新根目录 .gitignore**

追加前端相关忽略规则：

```
# Frontend
frontend/node_modules/
frontend/dist/
frontend/.umi/
frontend/.umi-production/
```

**Step 2: 更新 knowledge/capability.md**

追加：

```markdown
## 前端

- 框架：Ant Design Pro（umi max）
- UI：antd 5.x
- 开发地址：`http://localhost:8000`
- API 代理：`/api/*` → `http://localhost:8080`
```

**Step 3: Commit**

```bash
git add .gitignore backend/knowledge/capability.md
git commit -m "chore: update gitignore and docs for frontend"
```
