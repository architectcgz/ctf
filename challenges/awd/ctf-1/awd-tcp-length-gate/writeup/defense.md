# Defense Writeup

## 推荐修补位置

优先修改 `docker/workspace/src/challenge_app.py`，让 `CHECK` 继续工作，但不要再把当前 flag 暴露给外部请求。

## 修补原则

1. `PING` 必须继续返回 `PONG`。
2. `CHECK` 仍应保留原有业务入口和基本输入校验。
3. 任何异常分支、调试分支、长度门禁分支都不能回显真实 flag。

## 不应修改的内容

- `docker/runtime/app.py`
- `docker/runtime/ctf_runtime.py`
- `docker/check/check.py`
- `challenge.yml`

这些文件属于平台与 checker 契约边界，题目设计要求它们保持稳定。
