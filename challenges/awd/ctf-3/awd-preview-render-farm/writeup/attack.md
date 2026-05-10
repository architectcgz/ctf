# Attack Writeup

## 漏洞点

这题的公网入口 `render-web` 会把用户给的 `asset` 直接带给 `render-worker`。而 `render-worker` 会继续把它拼成：

```text
/internal/<asset>.json
```

它没有校验 `asset` 是否必须来自正常素材目录。

## 利用思路

1. 访问 `/catalog/demo`，确认正常素材路径类似：

```text
assets/receipt
```

2. 构造恶意素材路径：

```text
/preview?asset=debug/flag
```

3. `render-web` 把它交给 `render-worker`，后者继续访问：

```text
http://asset-cache:9092/internal/debug/flag.json
```

4. `asset-cache` 的调试接口直接返回当前动态 Flag。

## 关键边界

- `/api/flag` 是 checker 私有通道，不应作为选手攻击面。
- 攻击面应聚焦 `docker/workspace/src/challenge_app.py` 中素材路径的拼接逻辑。
- 这题需要 `render-web -> render-worker -> asset-cache` 的真实拓扑边界；压成单容器后，内部渲染流水线会失真。
