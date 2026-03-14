# challenge-pack-v1 Examples

这些目录用于演示 `challenge-pack-v1` 的作者侧源包形态，方便题目制作、审计、归档与后续实现导入器时做对照。

注意：

- 当前平台截至 2026-03-12 还没有公开的“challenge-pack Zip 一键导入”接口。
- 因此这里的示例是“源包示例”，不是“当前平台可直接上传并完整导入的交付包”。
- 示例中的 `runtime.image.ref` 表示最终运行镜像引用，`runtime.build.*` 表示可选的源码复现材料。

当前示例（未压缩形态）：

- `web-hello-01/`：最小 Web 容器题示例，用于演示题面、提示、动态 Flag 与镜像/构建材料写法
- `web-sqli-login-01/`：SQL 注入登录绕过示例，演示更接近真实 CTF Web 题的 manifest 组织方式

当前示例压缩包：

- `web-sqli-login-01.zip`：源包打包示例，用于说明目录与文件布局
