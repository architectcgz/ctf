# challenge-pack-v1 Examples

这些目录用于演示 `challenge-pack-v1` 题目包规范的“真实落地形态”，并按题目 `category` 做分类归档，方便教师/管理员参考与复用。

建议目录布局：

```text
ctf/docs/contracts/examples/challenge-pack-v1/
  web/
    <slug>/
    <slug>.zip
  pwn/
  reverse/
  crypto/
  misc/
  forensics/
```

当前示例（未压缩形态）：

- `web-hello-01/`：最小容器题示例（读取 `CTF_FLAG` 环境变量并回显，用于演示运行环境与 manifest 写法）
- `web-sqli-login-01/`：SQLi 登录绕过 + 读取 `secrets` 表（更接近真实 CTF Web 题）

当前示例（可直接上传的 Zip）：

- `web-sqli-login-01.zip`

