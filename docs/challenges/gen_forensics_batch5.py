#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

FORENSICS_CONFIGS = [
    {"slug": "forensics-blockchain", "title": "区块链取证", "difficulty": "hell", "tags": ["kp:blockchain"]},
    {"slug": "forensics-ai-deepfake", "title": "AI 深度伪造检测", "difficulty": "hell", "tags": ["kp:deepfake"]},
    {"slug": "forensics-quantum-crypto", "title": "量子密码分析", "difficulty": "hell", "tags": ["kp:quantum"]},
    {"slug": "forensics-stealth-malware", "title": "隐蔽恶意软件", "difficulty": "hell", "tags": ["kp:rootkit"]},
    {"slug": "forensics-anti-forensics", "title": "反取证技术", "difficulty": "hell", "tags": ["kp:anti-forensics"]},
    {"slug": "forensics-5g-network", "title": "5G 网络取证", "difficulty": "hard", "tags": ["kp:5g"]},
    {"slug": "forensics-satellite", "title": "卫星通信取证", "difficulty": "hard", "tags": ["kp:satellite"]},
    {"slug": "forensics-scada", "title": "SCADA 系统取证", "difficulty": "hard", "tags": ["kp:scada"]},
    {"slug": "forensics-drone", "title": "无人机取证", "difficulty": "hard", "tags": ["kp:drone"]},
    {"slug": "forensics-smart-contract", "title": "智能合约取证", "difficulty": "hard", "tags": ["kp:smart-contract"]},
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
    print(f"\n✅ Forensics 类别完成：50/50 题")
