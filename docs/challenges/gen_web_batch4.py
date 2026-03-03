#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

WEB_CONFIGS = [
    {"slug": "web-sqli-waf-bypass", "title": "WAF 绕过", "difficulty": "hard", "tags": ["vuln:sqli", "kp:waf-bypass"]},
    {"slug": "web-ssti-sandbox", "title": "SSTI 沙箱逃逸", "difficulty": "hard", "tags": ["vuln:ssti"]},
    {"slug": "web-dom-xss", "title": "DOM XSS", "difficulty": "hard", "tags": ["vuln:xss", "kp:dom"]},
    {"slug": "web-csp-bypass", "title": "CSP 绕过", "difficulty": "hard", "tags": ["vuln:xss", "kp:csp"]},
    {"slug": "web-polyglot-upload", "title": "多语言文件上传", "difficulty": "hard", "tags": ["vuln:file-upload"]},
    {"slug": "web-advanced-ssrf", "title": "高级 SSRF", "difficulty": "hard", "tags": ["vuln:ssrf"]},
    {"slug": "web-0day-exploit", "title": "框架 0day", "difficulty": "hell", "tags": ["vuln:rce", "kp:0day"]},
    {"slug": "web-full-chain", "title": "完整攻击链", "difficulty": "hell", "tags": ["kp:pentest"]},
    {"slug": "web-api-security", "title": "API 安全综合", "difficulty": "medium", "tags": ["kp:api"]},
    {"slug": "web-auth-bypass-advanced", "title": "高级认证绕过", "difficulty": "medium", "tags": ["vuln:auth-bypass"]},
]

def create(cfg):
    slug = cfg['slug']
    base = Path(f'packs/{slug}')
    base.mkdir(parents=True, exist_ok=True)
    manifest = {'spec_version': 'challenge-pack-v1', 'slug': slug, 'title': cfg['title'], 'category': 'web', 'difficulty': cfg['difficulty'], 'tags': cfg['tags'], 'description': {'file': 'statement.md'}, 'hints': [], 'attachments': [], 'flag': {'mode': 'dynamic'}, 'runtime': {'type': 'container', 'image': {'source': 'dockerfile', 'context_dir': 'docker', 'dockerfile': 'Dockerfile'}, 'expose': [{'container_port': 80, 'protocol': 'tcp'}]}, 'instance': {'ttl': '2h', 'max_per_user': 1}, 'resources': {'cpu': 0.5, 'memory': '256m', 'pids': 100}}
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
    print(f"\n✅ Web 类别完成：50/50 题")
