#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

MISC_CONFIGS = [
    {"slug": "misc-qr-code", "title": "二维码识别", "difficulty": "beginner", "tags": ["kp:qr-code"]},
    {"slug": "misc-barcode", "title": "条形码解析", "difficulty": "beginner", "tags": ["kp:barcode"]},
    {"slug": "misc-brainfuck", "title": "Brainfuck 解释", "difficulty": "beginner", "tags": ["kp:brainfuck"]},
    {"slug": "misc-piet", "title": "Piet 编程语言", "difficulty": "beginner", "tags": ["kp:piet"]},
    {"slug": "misc-whitespace", "title": "Whitespace 语言", "difficulty": "beginner", "tags": ["kp:whitespace"]},
    {"slug": "misc-esoteric-lang", "title": "深奥编程语言", "difficulty": "beginner", "tags": ["kp:esoteric"]},
    {"slug": "misc-unicode-trick", "title": "Unicode 技巧", "difficulty": "beginner", "tags": ["kp:unicode"]},
    {"slug": "misc-emoji-encode", "title": "Emoji 编码", "difficulty": "beginner", "tags": ["kp:emoji"]},
    {"slug": "misc-color-code", "title": "颜色编码", "difficulty": "beginner", "tags": ["kp:color"]},
    {"slug": "misc-music-notes", "title": "音符编码", "difficulty": "beginner", "tags": ["kp:music"]},
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
    print(f"\n已完成 Misc 类别：10/50 题")
