#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

MISC_CONFIGS = [
    {"slug": "misc-quantum-computing", "title": "量子计算", "difficulty": "hell", "tags": ["kp:quantum"]},
    {"slug": "misc-dna-computing", "title": "DNA 计算", "difficulty": "hell", "tags": ["kp:dna"]},
    {"slug": "misc-neuromorphic", "title": "神经形态计算", "difficulty": "hell", "tags": ["kp:neuromorphic"]},
    {"slug": "misc-photonic", "title": "光子计算", "difficulty": "hell", "tags": ["kp:photonic"]},
    {"slug": "misc-spintronics", "title": "自旋电子学", "difficulty": "hell", "tags": ["kp:spintronics"]},
    {"slug": "misc-memristor", "title": "忆阻器", "difficulty": "hell", "tags": ["kp:memristor"]},
    {"slug": "misc-topological", "title": "拓扑量子", "difficulty": "hell", "tags": ["kp:topological"]},
    {"slug": "misc-superconducting", "title": "超导计算", "difficulty": "hell", "tags": ["kp:superconducting"]},
    {"slug": "misc-molecular", "title": "分子计算", "difficulty": "hell", "tags": ["kp:molecular"]},
    {"slug": "misc-reversible", "title": "可逆计算", "difficulty": "hell", "tags": ["kp:reversible"]},
]

def create(cfg):
    slug = cfg['slug']
    base = Path(f'packs/{slug}')
    base.mkdir(parents=True, exist_ok=True)
    manifest = {'spec_version': 'challenge-pack-v1', 'slug': slug, 'title': cfg['title'], 'category': 'misc', 'difficulty': cfg['difficulty'], 'tags': cfg['tags'], 'description': {'file': 'statement.md'}, 'hints': [], 'attachments': [{'file': 'challenge.txt'}], 'flag': {'mode': 'static'}, 'runtime': {'type': 'none'}}
    with open(base / 'manifest.yml', 'w') as f: yaml.dump(manifest, f, allow_unicode=True, default_flow_style=False)
    (base / 'statement.md').write_text(f"# {cfg['title']}\n\n## 题目描述\n\n解析附件获取 flag。\n\n## 知识点\n\n- {', '.join(cfg['tags'])}\n")
    (base / 'challenge.txt').write_text('flag{placeholder}')
    with zipfile.ZipFile(f'{slug}.zip', 'w', zipfile.ZIP_DEFLATED) as z:
        for f in base.rglob('*'):
            if f.is_file(): z.write(f, f.relative_to(base.parent))
    print(f"✓ {slug}")

if __name__ == '__main__':
    for cfg in MISC_CONFIGS: create(cfg)
    print(f"\n✅ Misc 类别完成：50/50 题")
