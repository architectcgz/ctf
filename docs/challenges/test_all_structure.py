#!/usr/bin/env python3
import yaml
from pathlib import Path
from collections import defaultdict

stats = defaultdict(int)
errors = []

for manifest in sorted(Path('packs').rglob('manifest.yml')):
    try:
        with open(manifest) as f:
            data = yaml.safe_load(f)
            cat = data.get('category', 'unknown')
            stats[cat] += 1

            # 检查必填字段
            required = ['spec_version', 'slug', 'title', 'category', 'difficulty', 'tags', 'description', 'flag', 'runtime']
            missing = [k for k in required if k not in data]
            if missing:
                errors.append(f"{manifest.parent.name}: 缺少字段 {missing}")

            # 检查文件引用
            if data.get('description', {}).get('file'):
                desc_file = manifest.parent / data['description']['file']
                if not desc_file.exists():
                    errors.append(f"{manifest.parent.name}: statement.md 不存在")

    except Exception as e:
        errors.append(f"{manifest.parent.name}: {e}")

print("=== 题目统计 ===")
for cat, count in sorted(stats.items()):
    print(f"{cat}: {count} 题")
print(f"\n总计: {sum(stats.values())} 题")

if errors:
    print(f"\n=== 发现 {len(errors)} 个问题 ===")
    for e in errors[:20]:
        print(f"  ❌ {e}")
else:
    print("\n✅ 所有题目结构完整")
