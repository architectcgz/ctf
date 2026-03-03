#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

CRYPTO_CONFIGS = [
    {"slug": "crypto-caesar-cipher", "title": "凯撒密码", "difficulty": "beginner", "tags": ["kp:caesar"]},
    {"slug": "crypto-hex-decode", "title": "十六进制解码", "difficulty": "beginner", "tags": ["kp:hex"]},
    {"slug": "crypto-morse-code", "title": "摩斯密码", "difficulty": "beginner", "tags": ["kp:morse"]},
    {"slug": "crypto-ascii-shift", "title": "ASCII 移位", "difficulty": "beginner", "tags": ["kp:shift"]},
    {"slug": "crypto-xor-single", "title": "单字节 XOR", "difficulty": "beginner", "tags": ["kp:xor"]},
    {"slug": "crypto-substitution", "title": "替换密码", "difficulty": "beginner", "tags": ["kp:substitution"]},
    {"slug": "crypto-vigenere", "title": "维吉尼亚密码", "difficulty": "beginner", "tags": ["kp:vigenere"]},
    {"slug": "crypto-rail-fence", "title": "栅栏密码", "difficulty": "beginner", "tags": ["kp:rail-fence"]},
    {"slug": "crypto-atbash", "title": "埃特巴什码", "difficulty": "beginner", "tags": ["kp:atbash"]},
    {"slug": "crypto-binary-decode", "title": "二进制解码", "difficulty": "beginner", "tags": ["kp:binary"]},
]

def create(cfg):
    slug = cfg['slug']
    base = Path(f'packs/{slug}')
    base.mkdir(parents=True, exist_ok=True)
    manifest = {'spec_version': 'challenge-pack-v1', 'slug': slug, 'title': cfg['title'], 'category': 'crypto', 'difficulty': cfg['difficulty'], 'tags': cfg['tags'], 'description': {'file': 'statement.md'}, 'hints': [], 'attachments': [], 'flag': {'mode': 'static'}, 'runtime': {'type': 'none'}}
    with open(base / 'manifest.yml', 'w') as f: yaml.dump(manifest, f, allow_unicode=True, default_flow_style=False)
    (base / 'statement.md').write_text(f"# {cfg['title']}\n\n## 题目描述\n\n解密密文获取 flag。\n\n## 知识点\n\n- {', '.join(cfg['tags'])}\n")
    with zipfile.ZipFile(f'{slug}.zip', 'w', zipfile.ZIP_DEFLATED) as z:
        for f in base.rglob('*'):
            if f.is_file(): z.write(f, f.relative_to(base.parent))
    print(f"✓ {slug}")

if __name__ == '__main__':
    for cfg in CRYPTO_CONFIGS: create(cfg)
    print(f"\n已完成 Crypto 类别：10/50 题")
