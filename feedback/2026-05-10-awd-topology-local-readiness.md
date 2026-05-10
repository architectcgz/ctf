# AWD topology 本地就绪策略

## 问题描述

新增 AWD 三容器 topology 题时，本地验题最初使用：

```bash
docker compose up -d
python3 docker/check/check.py ...
```

这会出现一个稳定的竞态窗口：公网入口容器已经能接请求，但它依赖的内部容器还没完全 ready，结果本地 checker 的攻击链第一跳可能打到 `502` 或 `connection refused`。

## 原因分析

- `docker compose up -d` 只保证容器进入 `running`，不保证服务已经 ready
- 多容器题的攻击链往往是 `public -> internal-app -> internal-data`
- 只要 checker 覆盖真实 exploit 路径，就会放大这个启动顺序问题
- 单纯在 checker 里无限重试不是主方案，只能算兜底

## 解决方案

本项目内，AWD topology 题的本地调试和验题默认使用下面这组约定：

1. 每个服务都声明 `healthcheck`
2. 有依赖关系的服务使用 `depends_on.condition: service_healthy`
3. 本地启动优先使用：

```bash
docker compose up -d --build --wait --wait-timeout 60
```

4. `check/check.py` 可以保留小范围短重试，但只作为最后一层保险，不替代健康检查和就绪等待

## 收获

- 多容器题的“容器已启动”和“攻击链可验证”不是同一个状态
- 只测 SLA 的 checker 不容易暴露这个问题，但覆盖 exploit 的 checker 一定会撞上
- 这条规则应被视为 AWD topology 题的默认本地验证基线，而不是某一题的临时补丁
