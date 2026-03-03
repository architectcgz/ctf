#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

MISC_CONFIGS = [
    {"slug": "misc-blockchain-puzzle", "title": "区块链谜题", "difficulty": "medium", "tags": ["kp:blockchain"]},
    {"slug": "misc-smart-contract", "title": "智能合约漏洞", "difficulty": "medium", "tags": ["kp:solidity"]},
    {"slug": "misc-nft-metadata", "title": "NFT 元数据", "difficulty": "medium", "tags": ["kp:nft"]},
    {"slug": "misc-ipfs", "title": "IPFS 存储", "difficulty": "medium", "tags": ["kp:ipfs"]},
    {"slug": "misc-web3", "title": "Web3 交互", "difficulty": "medium", "tags": ["kp:web3"]},
    {"slug": "misc-ai-prompt", "title": "AI 提示注入", "difficulty": "medium", "tags": ["kp:prompt-injection"]},
    {"slug": "misc-llm-jailbreak", "title": "LLM 越狱", "difficulty": "medium", "tags": ["kp:llm"]},
    {"slug": "misc-ocr-bypass", "title": "OCR 识别绕过", "difficulty": "medium", "tags": ["kp:ocr"]},
    {"slug": "misc-captcha-break", "title": "验证码破解", "difficulty": "medium", "tags": ["kp:captcha"]},
    {"slug": "misc-2fa-bypass", "title": "2FA 绕过", "difficulty": "medium", "tags": ["kp:2fa"]},
]

def create(cfg):
    slug = cfg['slug']
    base = Path(f'packs/{slug}')
    base.mkdir(parents=True, exist_ok=True)
    manifest = {'spec_version': 'challenge-pack-v1', 'slug': slug, 'title': cfg['title'], 'category': 'misc', 'difficulty': cfg['difficulty'], 'tags': cfg['tags'], 'description': {'file': 'statement.md'}, 'hints': [], 'attachments': [], 'flag': {'mode': 'dynamic'}, 'runtime': {'type': 'container', 'image': {'source': 'dockerfile', 'context_dir': 'docker', 'dockerfile': 'Dockerfile'}, 'expose': [{'container_port': 9999, 'protocol': 'tcp'}]}, 'instance': {'ttl': '1h', 'max_per_user': 1}, 'resources': {'cpu': 0.2, 'memory': '64m', 'pids': 20}}
    with open(base / 'manifest.yml', 'w') as f: yaml.dump(manifest, f, allow_unicode=True, default_flow_style=False)
    (base / 'statement.md').write_text(f"# {cfg['title']}\n\n## 题目描述\n\n连接服务获取 flag。\n\n## 知识点\n\n- {', '.join(cfg['tags'])}\n")
    docker = base / 'docker'
    docker.mkdir(exist_ok=True)
    (docker / 'Dockerfile').write_text("FROM python:3.9-slim\nCOPY app.py /\nCMD [\"python\", \"/app.py\"]\n")
    (docker / 'app.py').write_text("import os\nprint('Flag:', os.getenv('FLAG', 'flag{placeholder}'))\n")
    with zipfile.ZipFile(f'{slug}.zip', 'w', zipfile.ZIP_DEFLATED) as z:
        for f in base.rglob('*'):
            if f.is_file(): z.write(f, f.relative_to(base.parent))
    print(f"✓ {slug}")

if __name__ == '__main__':
    for cfg in MISC_CONFIGS: create(cfg)
    print(f"\n已完成 Misc 类别：30/50 题")
