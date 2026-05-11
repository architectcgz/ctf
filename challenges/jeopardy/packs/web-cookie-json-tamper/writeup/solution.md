# 解法

Cookie 只是 Base64 包了一层 JSON，没有签名。把 `role` 改成 `admin` 后重新编码回去，请求 `/admin` 就会直接放行。
