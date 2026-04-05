# 前端工程目录规划

本文件用于约定下一步真正编写 Web 管理端时的目录结构。

当前阶段只做规划，不生成前端源码。


## 1. 建议目录结构

```text id="wxjv8q"
web/
├── ui-design.md
├── api-contract.md
├── frontend-structure.md
├── dist/
│   └── .gitkeep
└── app/
    ├── src/
    │   ├── main.tsx
    │   ├── App.tsx
    │   ├── api/
    │   │   ├── client.ts
    │   │   ├── config.ts
    │   │   ├── lists.ts
    │   │   ├── rules.ts
    │   │   └── render.ts
    │   ├── components/
    │   │   ├── layout/
    │   │   ├── common/
    │   │   ├── lists/
    │   │   ├── rules/
    │   │   └── render/
    │   ├── pages/
    │   │   ├── DashboardPage.tsx
    │   │   ├── ListsPage.tsx
    │   │   ├── RulesPage.tsx
    │   │   ├── RenderPage.tsx
    │   │   └── ConfigPage.tsx
    │   ├── types/
    │   │   └── api.ts
    │   ├── hooks/
    │   └── utils/
    └── package.json
```

---

## 2. 目录职责说明

### `api/`
只负责请求后端接口，不写页面逻辑。

例如：

- `config.ts`：读取完整配置
- `lists.ts`：list CRUD
- `rules.ts`：manual rule CRUD
- `render.ts`：触发渲染

### `components/`
只放可复用组件，不放页面级业务。

### `pages/`
按页面拆分：

- Dashboard
- Lists
- Rules
- Render
- Config

### `types/`
专门放前端 TypeScript 类型定义，尽量与后端字段保持同名。

---

## 3. 前端类型建议

建议在 `types/api.ts` 中定义这些类型：

- `ListDefinition`
- `ManualRule`
- `AppConfig`
- `RenderResponse`

字段名尽量和后端一致：

- `list_name`
- `priority`
- `enabled`
- `description`
- `output_path`

这样 API 层可以少做转换。

---

## 4. 前端页面开发顺序建议

1. `DashboardPage`
2. `ListsPage`
3. `RulesPage`
4. `RenderPage`
5. `ConfigPage`

原因：

- Dashboard 最容易先跑通
- Lists 和 Rules 是核心管理页面
- Render 依赖前面 API 已熟悉
- Config 页面最简单，收尾时加上即可

---

## 5. 当前阶段前后端分工边界

### 后端负责

- 配置合法性
- 地址合法性
- 去重与规范化
- 优先级覆盖
- diff / replace_all
- 配置持久化
- JSON 接口

### 前端负责

- 页面展示
- 表单交互
- 成功/失败提示
- 页面间导航
- 调用后端接口

前端不要重复实现后端业务规则。