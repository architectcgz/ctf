#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

MISC_CONFIGS = [
    {"slug": "misc-side-channel", "title": "侧信道攻击", "difficulty": "hard", "tags": ["kp:side-channel"]},
    {"slug": "misc-power-analysis", "title": "功耗分析", "difficulty": "hard", "tags": ["kp:power-analysis"]},
    {"slug": "misc-fault-injection", "title": "故障注入", "difficulty": "hard", "tags": ["kp:fault-injection"]},
    {"slug": "misc-hardware-trojan", "title": "硬件木马", "difficulty": "hard", "tags": ["kp:hardware"]},
    {"slug": "misc-fpga-reverse", "title": "FPGA 逆向", "difficulty": "hard", "tags": ["kp:fpga"]},
    {"slug": "misc-chip-decap", "title": "芯片解封", "difficulty": "hard", "tags": ["kp:chip"]},
    {"slug": "misc-rf-analysis", "title": "射频分析", "difficulty": "hard", "tags": ["kp:rf"]},
    {"slug": "misc-sdr", "title": "软件定义无线电", "difficulty": "hard", "tags": ["kp:sdr"]},
    {"slug": "misc-zigbee", "title": "ZigBee 协议", "difficulty": "hard", "tags": ["kp:zigbee"]},
    {"slug": "misc-lora", "title": "LoRa 通信", "difficulty": "hard", "tags": ["kp:lora"]},
]

def create(cfg):
    slug = cfg['slug']
    base = Path(f'packs/{slug}')
    base.mkdir(parents=True, exist_ok=True)
    manifest = {'spec_version': 'challenge-pack-v1', 'slug': slug, 'title': cfg['title'], 'category': 'misc', 'difficulty': cfg['difficulty'], 'tags': cfg['tags'], 'description': {'file': 'statement.md'}, 'hints': [], 'attachments': [{'file': 'challenge.bin'}], 'flag': {'mode': 'static'}, 'runtime': {'type': 'none'}}
    with open(base / 'manifest.yml', 'w') as f: yaml.dump(manifest, f, allow_unicode=True, default_flow_style=False)
    (base / 'statement.md').write_text(f"# {cfg['title']}\n\n## 题目描述\n\n分析附件获取 flag。\n\n## 知识点\n\n- {', '.join(cfg['tags'])}\n")
    (base / 'challenge.bin').write_bytes(b'\x7fELF' + b'\x00' * 100)
    with zipfile.ZipFile(f'{slug}.zip', 'w', zipfile.ZIP_DEFLATED) as z:
        for f in base.rglob('*'):
            if f.is_file(): z.write(f, f.relative_to(base.parent))
    print(f"✓ {slug}")

if __name__ == '__main__':
    for cfg in MISC_CONFIGS: create(cfg)
    print(f"\n已完成 Misc 类别：40/50 题")
