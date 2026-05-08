# Misc 题解：Invisible Sprint Briefing

## 解法

下载 `briefing.md` 后，肉眼看内容没有异常，但文件里夹带了零宽字符。题目约定使用：

- `U+200B` 表示二进制 `0`
- `U+200C` 表示二进制 `1`

下面的脚本可以直接抽出零宽字符并恢复字符串：

```bash
python3 - <<'PY'
from pathlib import Path

data = Path("briefing.md").read_text(encoding="utf-8")
bits = []
for ch in data:
    if ch == "\u200b":
        bits.append("0")
    elif ch == "\u200c":
        bits.append("1")

raw = "".join(bits)
out = []
for idx in range(0, len(raw), 8):
    chunk = raw[idx : idx + 8]
    if len(chunk) == 8:
        out.append(chr(int(chunk, 2)))
print("".join(out))
PY
```

得到兑换码：

```text
invisible-sprint-2026
```

然后请求：

```bash
curl 'http://127.0.0.1:18083/redeem?code=invisible-sprint-2026'
```

返回 JSON 中的 `flag` 即为答案。
