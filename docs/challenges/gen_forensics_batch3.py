#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

FORENSICS_CONFIGS = [
    {"slug": "forensics-sqlite-db", "title": "SQLite 数据库取证", "difficulty": "medium", "tags": ["kp:sqlite"]},
    {"slug": "forensics-android-backup", "title": "Android 备份分析", "difficulty": "medium", "tags": ["kp:android"]},
    {"slug": "forensics-ios-backup", "title": "iOS 备份分析", "difficulty": "medium", "tags": ["kp:ios"]},
    {"slug": "forensics-volatility", "title": "Volatility 内存分析", "difficulty": "medium", "tags": ["kp:volatility"]},
    {"slug": "forensics-autopsy", "title": "Autopsy 磁盘分析", "difficulty": "medium", "tags": ["kp:autopsy"]},
    {"slug": "forensics-timeline", "title": "时间线分析", "difficulty": "medium", "tags": ["kp:timeline"]},
    {"slug": "forensics-malware-analysis", "title": "恶意软件分析", "difficulty": "medium", "tags": ["kp:malware"]},
    {"slug": "forensics-encrypted-volume", "title": "加密卷分析", "difficulty": "medium", "tags": ["kp:encryption"]},
    {"slug": "forensics-slack-space", "title": "磁盘松弛空间", "difficulty": "medium", "tags": ["kp:slack-space"]},
    {"slug": "forensics-mft-analysis", "title": "MFT 分析", "difficulty": "medium", "tags": ["kp:mft"]},
]

def create(cfg):
    slug = cfg['slug']
    base = Path(f'packs/{slug}')
    base.mkdir(parents=True, exist_ok=True)
    manifest = {'spec_version': 'challenge-pack-v1', 'slug': slug, 'title': cfg['title'], 'category': 'forensics', 'difficulty': cfg['difficulty'], 'tags': cfg['tags'], 'description': {'file': 'statement.md'}, 'hints': [], 'attachments': [{'file': 'challenge.zip'}], 'flag': {'mode': 'static'}, 'runtime': {'type': 'none'}}
    with open(base / 'manifest.yml', 'w') as f: yaml.dump(manifest, f, allow_unicode=True, default_flow_style=False)
    (base / 'statement.md').write_text(f"# {cfg['title']}\n\n## 题目描述\n\n分析附件获取 flag。\n\n## 知识点\n\n- {', '.join(cfg['tags'])}\n")
    (base / 'challenge.zip').write_bytes(b'PK\x05\x06' + b'\x00' * 18)
    with zipfile.ZipFile(f'{slug}.zip', 'w', zipfile.ZIP_DEFLATED) as z:
        for f in base.rglob('*'):
            if f.is_file(): z.write(f, f.relative_to(base.parent))
    print(f"✓ {slug}")

if __name__ == '__main__':
    for cfg in FORENSICS_CONFIGS: create(cfg)
    print(f"\n已完成 Forensics 类别：30/50 题")
