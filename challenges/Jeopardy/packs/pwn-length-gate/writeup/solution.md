# Pwn 入门：长度门禁题解

## 解法

复制平台给出的 TCP 连接命令，连接后服务会提示输入一行内容：

```bash
nc <host> <port>
```

普通输入会得到长度回显：

```text
hello
```

返回会提示需要更长输入和 magic word。构造超过 40 字节且包含 `magic` 的 payload 即可：

```bash
python3 - <<'PY' | nc <host> <port>
print('A' * 44 + 'magic')
PY
```

服务会返回平台为当前实例生成的动态 flag，形如：

```text
flag{...}
```

复制服务返回的完整 flag 到平台提交即可。
