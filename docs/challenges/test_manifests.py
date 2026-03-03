#!/usr/bin/env python3
import yaml
from pathlib import Path

errors = []
for manifest in Path('packs').rglob('manifest.yml'):
    try:
        with open(manifest) as f:
            data = yaml.safe_load(f)
            if not data.get('slug') or not data.get('title'):
                errors.append(f"{manifest}: 缺少必填字段")
    except Exception as e:
        errors.append(f"{manifest}: {e}")

if errors:
    for e in errors[:10]: print(e)
    print(f"\n总计 {len(errors)} 个错误")
else:
    print("✅ 所有 manifest.yml 格式正确")
