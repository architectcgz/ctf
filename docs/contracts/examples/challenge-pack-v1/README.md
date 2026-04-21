# challenge-pack-v1 Examples

这些目录用于演示 `challenge-pack-v1` 的作者侧源包形态，方便题目制作、审计、归档与后续实现导入器时做对照。

注意：

- 当前平台截至 2026-03-31 已支持后台 `challenge.yml` 题目包上传预览与确认导入，以及复用同一解析逻辑的 CLI 导入。
- 因此这里的示例统一以 `challenge.yml` 作为题目清单。
- 示例中的 `runtime.image.ref` 表示最终运行镜像引用，`docker/` 目录仍可作为源码复现与审计材料保留。

当前示例（未压缩形态）：

- `web-hello-01/`：最小 Web 容器题示例，用于演示题面、提示、动态 Flag 与镜像/构建材料写法
- `web-sqli-login-01/`：SQL 注入登录绕过示例，演示更接近真实 CTF Web 题的 `challenge.yml` 组织方式
- `awd-bank-portal-01/`：最小可用 AWD 服务模板示例，演示 `meta.mode: awd` 与 `extensions.awd` 的模板定义方式

当前示例压缩包：

- `web-sqli-login-01.zip`：源包打包示例，用于说明目录与文件布局
- `awd-bank-portal-01.zip`：AWD 服务模板包打包示例，可直接用于后台 AWD 模板导入预览与确认导入
