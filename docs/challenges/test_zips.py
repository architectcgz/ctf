#!/usr/bin/env python3
import zipfile
from pathlib import Path

errors = []
for zf in Path('.').glob('*.zip'):
    try:
        with zipfile.ZipFile(zf, 'r') as z:
            if z.testzip() is not None:
                errors.append(f"{zf}: ZIP 损坏")
    except Exception as e:
        errors.append(f"{zf}: {e}")

if errors:
    for e in errors[:10]: print(e)
    print(f"\n总计 {len(errors)} 个错误")
else:
    print(f"✅ 所有 ZIP 包完整（共 {len(list(Path('.').glob('*.zip')))} 个）")
