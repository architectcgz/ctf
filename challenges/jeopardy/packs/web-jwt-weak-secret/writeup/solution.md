# 解法

先拿 guest token 观察结构，再用弱密钥 `changeme123` 自己重签一份 `role=admin` 的 payload。带着它请求 `/admin` 就能拿到 flag。
