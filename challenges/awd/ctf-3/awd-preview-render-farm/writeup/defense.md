# 防守题解

## 可修改位置

- 平台防守入口对应的业务代码在 `/workspace/src/challenge_app.py`
- 题目包内源文件位置是 `docker/workspace/src/challenge_app.py`
- 受保护的 `runtime` 与 `check` 代码不要动

## 主要漏洞点

### 1. render-worker 不该信任外部传入的原始 asset 路径

当前实现把 `asset` 直接拼到内部路径上，导致攻击者可以越过正常素材目录，命中调试接口。

推荐修法：

- 只允许固定素材 key，例如 `receipt`、`label`
- 在 worker 内部维护 allowlist，而不是接受任意 path
- 入口层和 worker 层都做一次最小必要校验

### 2. asset-cache 的调试接口不该暴露 Flag

`/internal/debug/flag.json` 直接把 Flag 当作调试内容输出，一旦上游路径校验失手就会直接失陷。

推荐修法：

- 删除 `flag` 字段
- 或改成固定占位值 / 统计摘要
- 如果必须保留调试接口，至少增加服务间鉴权

## 推荐修改函数 / 角色

- `preview()`
- `internal_render()`
- `debug_flag()`

## 保活约束

- `/health` 正常返回 200
- `/api/flag` 继续支持 checker 的 PUT / GET
- `/catalog/demo` 仍然可用
- `/preview` 仍可渲染合法素材

## 交付判断

1. 恶意 asset 路径不能再命中内部调试接口
2. 调试接口不再直接回显 Flag
3. `/health`、`/api/flag`、`/catalog/demo` 仍保持可用
