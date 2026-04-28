# Crypto 入门：移位明信片题解

## 解法

附件给出 `shift = 7`，密文是：

```text
mshn{jyfwav_jhlzhy_wvzajhyk}
```

Caesar 加密向后移动 7 位，解密时每个英文字母向前移动 7 位：

```python
s = "mshn{jyfwav_jhlzhy_wvzajhyk}"
out = ""
for ch in s:
    if "a" <= ch <= "z":
        out += chr((ord(ch) - ord("a") - 7) % 26 + ord("a"))
    else:
        out += ch
print(out)
```

得到 flag：

```text
flag{crypto_caesar_postcard}
```
