# 攻击思路

## 路径穿越

预览接口直接使用 `path` 参数拼接 `/data/uploads`：

```text
/preview?path=../../flag
```

未修复时可读取动态 Flag。

## 上传绕过

上传逻辑错误允许 `.jpg.php`，真实部署到可执行 PHP 环境时会形成 WebShell 风险。本演示包使用 Flask，不执行 PHP，只用于说明绕过点。
