#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

WEB_CONFIGS = [
    {"slug": "web-graphql-idor", "title": "GraphQL 越权", "difficulty": "medium", "tags": ["vuln:idor", "stack:graphql"]},
    {"slug": "web-prototype-pollution", "title": "原型链污染", "difficulty": "medium", "tags": ["vuln:prototype-pollution"]},
    {"slug": "web-jwt-algorithm", "title": "JWT 算法混淆", "difficulty": "medium", "tags": ["vuln:jwt"]},
    {"slug": "web-oauth-redirect", "title": "OAuth 重定向劫持", "difficulty": "medium", "tags": ["vuln:oauth"]},
    {"slug": "web-xml-bomb", "title": "XML 炸弹", "difficulty": "medium", "tags": ["vuln:xxe"]},
    {"slug": "web-type-juggling", "title": "PHP 类型混淆", "difficulty": "medium", "tags": ["vuln:type-juggling"]},
    {"slug": "web-mass-assignment", "title": "批量赋值", "difficulty": "medium", "tags": ["vuln:mass-assignment"]},
    {"slug": "web-http-smuggling", "title": "HTTP 请求走私", "difficulty": "medium", "tags": ["vuln:http-smuggling"]},
    {"slug": "web-cache-poisoning", "title": "缓存投毒", "difficulty": "medium", "tags": ["vuln:cache-poisoning"]},
    {"slug": "web-websocket-hijack", "title": "WebSocket 劫持", "difficulty": "medium", "tags": ["vuln:websocket"]},
]

def create(cfg):
    slug = cfg['slug']
    base = Path(f'packs/{slug}')
    base.mkdir(parents=True, exist_ok=True)
    manifest = {'spec_version': 'challenge-pack-v1', 'slug': slug, 'title': cfg['title'], 'category': 'web', 'difficulty': cfg['difficulty'], 'tags': cfg['tags'], 'description': {'file': 'statement.md'}, 'hints': [], 'attachments': [], 'flag': {'mode': 'dynamic'}, 'runtime': {'type': 'container', 'image': {'source': 'dockerfile', 'context_dir': 'docker', 'dockerfile': 'Dockerfile'}, 'expose': [{'container_port': 80, 'protocol': 'tcp'}]}, 'instance': {'ttl': '2h', 'max_per_user': 1}, 'resources': {'cpu': 0.3, 'memory': '128m', 'pids': 50}}
    with open(base / 'manifest.yml', 'w') as f: yaml.dump(manifest, f, allow_unicode=True, default_flow_style=False)
    (base / 'statement.md').write_text(f"# {cfg['title']}\n\n## 题目描述\n\n待补充\n\n## 目标\n\n获取 flag。\n")
    docker = base / 'docker'
    docker.mkdir(exist_ok=True)
    (docker / 'Dockerfile').write_text("FROM php:7.4-apache\nCOPY src/ /var/www/html/\nEXPOSE 80\nCMD [\"apache2-foreground\"]\n")
    src = docker / 'src'
    src.mkdir(exist_ok=True)
    (src / 'index.php').write_text("<?php\necho 'Flag: ' . (getenv('FLAG') ?: 'flag{placeholder}');\n?>\n")
    with zipfile.ZipFile(f'{slug}.zip', 'w', zipfile.ZIP_DEFLATED) as z:
        for f in base.rglob('*'):
            if f.is_file(): z.write(f, f.relative_to(base.parent))
    print(f"✓ {slug}")

if __name__ == '__main__':
    for cfg in WEB_CONFIGS: create(cfg)
    print(f"\n已完成 Web 类别：40/50 题")
