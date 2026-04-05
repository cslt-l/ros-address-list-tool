# Web 管理端设计稿

## 1. 设计目标

Web 管理端不是一个“在线改大文本脚本”的页面，而是一个围绕配置与渲染流程构建的小型管理台。

它的核心目标：

1. 管理 address-list 定义
2. 管理 description
3. 管理 manual rules
4. 触发渲染
5. 查看渲染结果
6. 查看当前配置快照
7. 为后续日志查看、历史记录、RouterOS 实机同步预留空间

---

## 2. 页面划分

第一版建议拆成 5 个页面。

### 2.1 总览页（Dashboard）

用途：

- 展示当前系统状态
- 展示当前配置摘要
- 展示最近一次渲染入口
- 作为进入其它页面的导航首页

建议显示内容：

- 当前 address-list 数量
- 当前 manual rule 数量
- 当前默认渲染模式
- 当前输出路径
- 当前 HTTP 服务状态
- 快速入口按钮：
    - 进入 list 管理
    - 进入 rule 管理
    - 立即渲染
    - 查看原始配置

---

### 2.2 Address List 管理页

用途：

- 管理所有 list 定义
- 新增 / 编辑 / 删除 list
- 单独修改 description

页面结构建议：

- 顶部：新增 list 按钮
- 中间：list 表格
- 右侧或弹窗：编辑表单

表格字段建议：

- name
- family
- enabled
- description
- 操作按钮：
    - 查看
    - 编辑
    - 删除
    - 修改 description

编辑表单字段建议：

- name
- family（ipv4 / ipv6）
- enabled
- description

---

### 2.3 Manual Rule 管理页

用途：

- 管理所有手工规则
- 新增 / 编辑 / 删除规则

表格字段建议：

- id
- list_name
- action
- priority
- enabled
- description
- entries 数量
- 操作按钮：
    - 编辑
    - 删除

编辑表单字段建议：

- id（新增时可空，后端自动生成）
- list_name
- action（add / remove）
- priority
- enabled
- description
- entries（多行输入）

---

### 2.4 渲染中心页

用途：

- 触发渲染
- 预览结果
- 查看脚本输出
- 后续可扩展为“历史记录中心”

建议布局：

- 左侧：渲染参数
- 右侧：脚本结果预览

渲染参数建议：

- mode（replace_all / diff）
- output_path（可选临时覆盖）
- 执行按钮

渲染结果建议：

- 返回 mode
- list_count
- entry_count
- output_path
- script 文本框
- 复制脚本按钮

---

### 2.5 原始配置查看页

用途：

- 直接查看当前后端配置对象
- 作为排查页面使用
- 帮助开发阶段确认前端和后端字段一致

建议显示：

- `/api/v1/config` 的原始 JSON
- 支持折叠展开
- 支持复制

---

## 3. 顶部导航建议

建议顶部导航栏包含：

- 总览
- Address Lists
- Manual Rules
- Render
- Config

这样第一版足够清晰，不需要做过多菜单层级。

---

## 4. 页面跳转建议

建议的前端路由：

- `/`
- `/lists`
- `/rules`
- `/render`
- `/config`

后续如果要加详情页，可以扩展：

- `/lists/:name`
- `/rules/:id`

---

## 5. 交互原则

### 5.1 所有“修改型操作”都要有明确反馈

例如：

- 新增成功
- 更新成功
- 删除成功
- 渲染成功
- 请求失败

建议使用统一消息提示区，不要每个页面自己发散处理。

### 5.2 删除操作要二次确认

尤其是：

- 删除 list
- 删除 manual rule

因为这些操作会直接影响配置持久化。

### 5.3 表单字段尽量和后端字段同名

例如：

- `list_name`
- `priority`
- `enabled`
- `description`

这样前后端对接最省事。

---

## 6. 第一版暂时不做的内容

为了控制复杂度，第一版前端先不做：

- 用户登录
- 权限系统
- 多用户协作
- 操作审计页
- 历史渲染记录页
- 日志查看页
- RouterOS 在线同步页

这些内容以后都可以在当前结构上继续加。

---

## 7. 与当前后端接口的对应关系

### Address List 页对应接口

- `GET /api/v1/lists`
- `POST /api/v1/lists`
- `GET /api/v1/lists/{name}`
- `PUT /api/v1/lists/{name}`
- `DELETE /api/v1/lists/{name}`
- `GET /api/v1/lists/{name}/description`
- `PUT /api/v1/lists/{name}/description`

### Manual Rule 页对应接口

- `GET /api/v1/manual-rules`
- `POST /api/v1/manual-rules`
- `PUT /api/v1/manual-rules/{id}`
- `DELETE /api/v1/manual-rules/{id}`

### Render 页对应接口

- `POST /api/v1/render`

### Config 页对应接口

- `GET /api/v1/config`

---

## 8. 当前已知注意事项

### 8.1 Windows / PowerShell 中文编码

开发阶段已经观察到：

- PowerShell 直接查看 UTF-8 文件时可能出现乱码
- 交互式输入中文 JSON 时，可能出现 `?`

因此前端阶段应尽量避免依赖手工终端输入中文，而应通过浏览器表单提交 JSON。

### 8.2 前端不要自己实现业务规则

例如：

- 不要在前端自己判断优先级覆盖规则
- 不要在前端自己计算 diff
- 不要在前端自己规范化 IP/CIDR

这些都应该以后端返回结果为准。

---

## 9. 下一步前端实现顺序建议

1. 搭前端工程骨架
2. 先接 Dashboard
3. 再接 Address List 管理页
4. 再接 Manual Rule 管理页
5. 最后接 Render 页和 Config 页