#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

FORENSICS_CONFIGS = [
    {"slug": "forensics-ntfs-ads", "title": "NTFS 交替数据流", "difficulty": "hard", "tags": ["kp:ads"]},
    {"slug": "forensics-vmdk-analysis", "title": "VMDK 虚拟磁盘", "difficulty": "hard", "tags": ["kp:vmdk"]},
    {"slug": "forensics-bitlocker", "title": "BitLocker 解密", "difficulty": "hard", "tags": ["kp:bitlocker"]},
    {"slug": "forensics-apfs", "title": "APFS 文件系统", "difficulty": "hard", "tags": ["kp:apfs"]},
    {"slug": "forensics-ext4", "title": "EXT4 文件系统", "difficulty": "hard", "tags": ["kp:ext4"]},
    {"slug": "forensics-raid-recovery", "title": "RAID 恢复", "difficulty": "hard", "tags": ["kp:raid"]},
    {"slug": "forensics-firmware-extract", "title": "固件提取", "difficulty": "hard", "tags": ["kp:firmware"]},
    {"slug": "forensics-car-forensics", "title": "车载系统取证", "difficulty": "hard", "tags": ["kp:automotive"]},
    {"slug": "forensics-iot-forensics", "title": "IoT 设备取证", "difficulty": "hard", "tags": ["kp:iot"]},
    {"slug": "forensics-cloud-forensics", "title": "云取证", "difficulty": "hard", "tags": ["kp:cloud"]},
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
    print(f"\n已完成 Forensics 类别：40/50 题")
