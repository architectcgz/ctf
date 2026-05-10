# Attack Writeup

## 漏洞点

这题的公网入口 `patch-gateway` 允许客户端传入 `signer_role`，然后直接把它变成内部请求头 `X-Signer-Role` 交给 `signer-core`。而 `signer-core` 会把这个头当成可信岗位权限。

## 利用思路

1. 访问 `/channels/demo`，确认公网入口支持生成 `stable` 渠道的 bundle。
2. 构造请求：

```text
/bundle?channel=stable&signer_role=release-manager
```

3. `patch-gateway` 会把 `signer_role` 变成 `X-Signer-Role: release-manager`。
4. `signer-core` 误以为请求来自发布岗位，于是改走高权限 `keyset` 链路。
5. `key-vault` 返回的高权限结果里直接包含当前动态 Flag。

## 关键边界

- `/api/flag` 是 checker 私有通道，不应作为选手攻击面。
- 攻击面应聚焦 `docker/workspace/src/challenge_app.py` 中能力角色如何从公网穿透到 signer。
- 这题需要 `patch-gateway -> signer-core -> key-vault` 的真实拓扑边界；压成单容器后，权限边界会失真。
