# Forensics 入门：回收站便签题解

## 解法

先列出归档内容：

```bash
tar -tf evidence.tar
```

可以看到隐藏目录：

```text
evidence/.trash/deleted-note.txt
```

读取该文件：

```bash
tar -xOf evidence.tar evidence/.trash/deleted-note.txt
```

得到 flag：

```text
flag{forensics_trash_note}
```
