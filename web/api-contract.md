# Web 管理端 API 对接契约

本文件用于约定当前后端接口的请求与响应结构，供前端直接对接。

---

## 1. 通用约定

### 1.1 返回格式

当前后端大多数接口直接返回 JSON 对象或 JSON 数组。

成功时：

- 返回业务对象
- 或返回 `{ "message": "..." }`

失败时：

- 返回 `{ "error": "..." }`

### 1.2 Content-Type

前端发送 JSON 请求体时，统一使用：

```text id="kf4c6q"
application/json; charset=utf-8
```
## 2. 健康检查

### GET `/healthz`

返回示例：

```json id="07z2tb"
{
  "status": "ok"
}
```

---

## 3. 获取完整配置

### GET `/api/v1/config`

返回示例字段：

- `auto_create_lists`
- `log_file`
- `lists`
- `desired_sources`
- `current_state_sources`
- `manual_rules`
- `output`
- `server`

前端用途：

- Config 页面直接展示
- Dashboard 页面做摘要统计

---

## 4. 渲染接口

### POST `/api/v1/render`

请求体可为空。

可选请求体：

```json id="yjkm18"
{
  "mode": "diff",
  "output_path": "./output/http-diff.rsc"
}
```

响应示例：

```json id="n3kghl"
{
  "mode": "diff",
  "list_count": 6,
  "entry_count": 16,
  "output_path": "./output/http-diff.rsc",
  "script": "/ip firewall address-list\n..."
}
```

前端用途：

- Render 页面主接口

---

## 5. Address List 接口

### GET `/api/v1/lists`

返回数组，每个元素结构为：

```json id="xvjlwm"
{
  "name": "toWanTelecom",
  "family": "ipv4",
  "enabled": true,
  "description": "电信出口地址列表"
}
```

---

### POST `/api/v1/lists`

请求体：

```json id="wx6mym"
{
  "name": "toWanIpv4",
  "family": "ipv4",
  "enabled": true,
  "description": "新增的 IPv4 出口列表"
}
```

响应：

```json id="8lfslu"
{
  "message": "list upserted"
}
```

---

### GET `/api/v1/lists/{name}`

响应：

```json id="i50i5y"
{
  "name": "toWanTelecom",
  "family": "ipv4",
  "enabled": true,
  "description": "电信出口地址列表"
}
```

---

### PUT `/api/v1/lists/{name}`

请求体：

```json id="hlmvyk"
{
  "family": "ipv4",
  "enabled": true,
  "description": "更新后的 IPv4 列表说明"
}
```

注意：

- 路径里的 `{name}` 优先生效
- 请求体中的 `name` 不作为最终依据

响应：

```json id="608do9"
{
  "message": "list updated"
}
```

---

### DELETE `/api/v1/lists/{name}`

响应：

```json id="cnspu7"
{
  "message": "list deleted"
}
```

---

### GET `/api/v1/lists/{name}/description`

响应：

```json id="s3jlwm"
{
  "name": "toWanTelecom",
  "description": "电信出口地址列表"
}
```

---

### PUT `/api/v1/lists/{name}/description`

请求体：

```json id="t7qy1t"
{
  "description": "只改 description 的版本"
}
```

响应：

```json id="a90ih9"
{
  "message": "description updated"
}
```

---

## 6. Manual Rule 接口

### GET `/api/v1/manual-rules`

返回数组，每个元素结构为：

```json id="o8sd0s"
{
  "id": "rule-telecom-add",
  "list_name": "toWanTelecom",
  "action": "add",
  "priority": 1000,
  "enabled": true,
  "description": "手工强制加入电信地址",
  "entries": [
    "223.5.5.5",
    "223.6.6.6"
  ]
}
```

---

### POST `/api/v1/manual-rules`

请求体示例：

```json id="e13w2w"
{
  "id": "rule-temp-balance-add",
  "list_name": "toPccBalance",
  "action": "add",
  "priority": 1500,
  "enabled": true,
  "description": "临时加入一个测试地址",
  "entries": [
    "10.10.10.10"
  ]
}
```

响应：

```json id="hi781l"
{
  "message": "manual rule upserted"
}
```

---

### PUT `/api/v1/manual-rules/{id}`

请求体示例：

```json id="b25asi"
{
  "list_name": "toPccBalance",
  "action": "add",
  "priority": 1600,
  "enabled": true,
  "description": "更新后的测试规则",
  "entries": [
    "10.10.10.10",
    "10.10.10.11"
  ]
}
```

响应：

```json id="mjlwmn"
{
  "message": "manual rule updated"
}
```

---

### DELETE `/api/v1/manual-rules/{id}`

响应：

```json id="oj8p7g"
{
  "message": "manual rule deleted"
}
```

---

## 7. 前端状态建议

### 服务端状态（Server State）

这些数据建议通过接口实时获取：

- lists
- manual_rules
- config
- render result

### 本地表单状态（Local UI State）

这些数据建议只存在页面本地：

- 当前弹窗是否打开
- 表单输入框内容
- 当前选中的 list / rule
- 当前渲染表单参数

---

## 8. 前端错误处理建议

后端失败时通常返回：

```json id="ycs8xr"
{
  "error": "..."
}
```

前端统一处理方式建议:
1. 优先展示 `error`
2. 没有 `error` 时展示通用失败消息
3. 删除操作失败时不要自动刷新页面
4. 新增/更新成功后再刷新列表
