# 解法

直接在抓包里找 `Authorization: Basic ...`。把 Base64 解开后会得到 `username:password`，本题把 flag 放在 password 位置。
