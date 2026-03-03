#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

WEB_CONFIGS = [
    {"slug": "web-idor", "title": "越权访问", "difficulty": "easy", "tags": ["vuln:idor"]},
    {"slug": "web-jwt-weak", "title": "JWT 弱密钥", "difficulty": "easy", "tags": ["vuln:jwt"]},
    {"slug": "web-ssti-basic", "title": "服务端模板注入", "difficulty": "easy", "tags": ["vuln:ssti"]},
    {"slug": "web-ssrf-basic", "title": "SSRF", "difficulty": "easy", "tags": ["vuln:ssrf"]},
    {"slug": "web-deserialization", "title": "反序列化", "difficulty": "easy", "tags": ["vuln:deserialization"]},
    {"slug": "web-race-condition", "title": "条件竞争", "difficulty": "easy", "tags": ["vuln:race-condition"]},
    {"slug": "web-open-redirect", "title": "开放重定向", "difficulty": "easy", "tags": ["vuln:open-redirect"]},
    {"slug": "web-path-traversal", "title": "路径遍历", "difficulty": "easy", "tags": ["vuln:path-traversal"]},
    {"slug": "web-sqli-blind", "title": "盲注", "difficulty": "medium", "tags": ["vuln:sqli", "kp:blind"]},
    {"slug": "web-nosql-injection", "title": "NoSQL 注入", "difficulty": "medium", "tags": ["vuln:nosql"]},
]

def create(cfg):
    slug = cfg['slug']
    base = Path(f'packs/{slug}')
    base.mkdir(parents=True, exist_ok=True)

    manifest = {
        'spec_version': 'challenge-pack-v1', 'slug': slug, 'title': cfg['title'],
        'category': 'web', 'difficulty': cfg['difficulty'], 'tags': cfg['tags'],
        'description': {'file': 'statement.md'}, 'hints': [], 'attachments': [],
        'flag': {'mode': 'dynamic'},
        'runtime': {'type': 'container', 'image': {'source': 'dockerfile', 'context_dir': 'docker', 'dockerfile': 'Dockerfile'},
                   'expose': [{'container_port': 80, 'protocol': 'tcp'}]},
        'instance': {'ttl': '2h', 'max_per_user': 1},
        'resources': {'cpu': 0.3, 'memory': '128m', 'pids': 50}
    }

    with open(base / 'manifest.yml', 'w') as f:
        yaml.dump(manifest, f, allow_unicode=True, default_flow_style=False)

    (base / 'statement.md').write_text(f"# {cfg['title']}\n\n## 题目描述\n\n待补充\n\n## 目标\n\n获取 flag。\n")

    docker = base / 'docker'
    docker.mkdir(exist_ok=True)
    (docker / 'Dockerfile').write_text("FROM php:7.4-apache\nCOPY src/ /var/www/html/\nEXPOSE 80\nCMD [\"apache2-foreground\"]\n")

    src = docker / 'src'
    src.mkdir(exist_ok=True)
    (src / 'index.php').write_text("<?php\necho 'Flag: ' . (getenv('FLAG') ?: 'flag{placeholder}');\n?>\n")

    with zipfile.ZipFile(f'{slug}.zip', 'w', zipfile.ZIP_DEFLATED) as z:
        for f in base.rglob('*'):
            if f.is_file():
                z.write(f, f.relative_to(base.parent))

    print(f"✓ {slug}")

if __name__ == '__main__':
    for cfg in WEB_CONFIGS:
        create(cfg)
    print(f"\n已完成 Web 类别：30/50 题")
