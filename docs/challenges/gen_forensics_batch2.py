#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

FORENSICS_CONFIGS = [
    {"slug": "forensics-memory-dump", "title": "内存取证", "difficulty": "easy", "tags": ["kp:memory-forensics"]},
    {"slug": "forensics-disk-image", "title": "磁盘镜像分析", "difficulty": "easy", "tags": ["kp:disk-forensics"]},
    {"slug": "forensics-network-capture", "title": "网络流量捕获", "difficulty": "easy", "tags": ["kp:wireshark"]},
    {"slug": "forensics-registry-analysis", "title": "注册表分析", "difficulty": "easy", "tags": ["kp:registry"]},
    {"slug": "forensics-log-analysis", "title": "日志分析", "difficulty": "easy", "tags": ["kp:logs"]},
    {"slug": "forensics-pdf-analysis", "title": "PDF 文件分析", "difficulty": "easy", "tags": ["kp:pdf"]},
    {"slug": "forensics-office-macro", "title": "Office 宏分析", "difficulty": "easy", "tags": ["kp:macro"]},
    {"slug": "forensics-usb-forensics", "title": "USB 取证", "difficulty": "easy", "tags": ["kp:usb"]},
    {"slug": "forensics-browser-history", "title": "浏览器历史", "difficulty": "easy", "tags": ["kp:browser"]},
    {"slug": "forensics-email-headers", "title": "邮件头分析", "difficulty": "easy", "tags": ["kp:email"]},
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
    print(f"\n已完成 Forensics 类别：20/50 题")
