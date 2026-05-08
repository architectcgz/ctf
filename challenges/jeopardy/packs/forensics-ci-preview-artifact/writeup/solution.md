# Forensics 题解：CI Preview Artifact Leak

## 解法

先把工件包里的文本内容扫一遍：

```bash
unzip -l preview-bundle.zip
unzip -p preview-bundle.zip logs/runner.log
unzip -p preview-bundle.zip reports/step-summary.txt
```

可以在日志中看到一段 Base64 文本，例如：

```text
preview token fragment saved as: cHJldmlldy1hcnRpZmFjdC0yMDI2LXRva2Vu
```

直接解码即可：

```bash
python3 - <<'PY'
import base64
print(base64.b64decode("cHJldmlldy1hcnRpZmFjdC0yMDI2LXRva2Vu").decode())
PY
```

得到 token：

```text
preview-artifact-2026-token
```

然后访问：

```bash
curl 'http://127.0.0.1:18084/redeem?token=preview-artifact-2026-token'
```

返回 JSON 中的 `flag` 即为答案。
