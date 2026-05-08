# Crypto 题解：Stream Backup Ticket

## 解法

题面给了一份已知明文和对应密文，这意味着可以直接恢复整段 keystream：

```python
known_plain = b"backup-ticket:readonly"
known_cipher = bytes.fromhex("267045194f7fbd57387ae872137ed35db378661ed853")
keystream = bytes(a ^ b for a, b in zip(known_plain, known_cipher))
print(keystream.hex())
```

再用同一段 keystream 去解第二份密文：

```bash
python3 - <<'PY'
known_plain = b"backup-ticket:readonly"
known_cipher = bytes.fromhex("267045194f7fbd57387ae872137ed35db378661ed853")
secret_cipher = bytes.fromhex("376554175b62bd11612bb53a0525c253a76c")
keystream = bytes(a ^ b for a, b in zip(known_plain, known_cipher))
code = bytes(a ^ b for a, b in zip(secret_cipher, keystream))
print(code.decode())
PY
```

得到恢复码：

```text
stream-2026-backup
```

最后兑换：

```bash
curl 'http://127.0.0.1:18082/redeem?code=stream-2026-backup'
```

返回 JSON 中的 `flag` 即为答案。
