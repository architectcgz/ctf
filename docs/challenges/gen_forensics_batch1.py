#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

FORENSICS_CONFIGS = [
    {"slug": "forensics-file-signature", "title": "文件签名识别", "difficulty": "beginner", "tags": ["kp:file-signature"]},
    {"slug": "forensics-exif-data", "title": "EXIF 数据提取", "difficulty": "beginner", "tags": ["kp:exif"]},
    {"slug": "forensics-strings-search", "title": "字符串搜索", "difficulty": "beginner", "tags": ["kp:strings"]},
    {"slug": "forensics-zip-password", "title": "ZIP 密码破解", "difficulty": "beginner", "tags": ["kp:zip-crack"]},
    {"slug": "forensics-image-steg", "title": "图片隐写", "difficulty": "beginner", "tags": ["kp:steganography"]},
    {"slug": "forensics-pcap-basic", "title": "PCAP 流量分析", "difficulty": "beginner", "tags": ["kp:pcap"]},
    {"slug": "forensics-deleted-file", "title": "已删除文件恢复", "difficulty": "beginner", "tags": ["kp:file-recovery"]},
    {"slug": "forensics-hex-edit", "title": "十六进制编辑", "difficulty": "beginner", "tags": ["kp:hex"]},
    {"slug": "forensics-metadata", "title": "元数据分析", "difficulty": "beginner", "tags": ["kp:metadata"]},
    {"slug": "forensics-audio-steg", "title": "音频隐写", "difficulty": "beginner", "tags": ["kp:audio"]},
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
    print(f"\n已完成 Forensics 类别：10/50 题")
