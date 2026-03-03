#!/usr/bin/env python3
"""
CTF 题目批量生成脚本
快速生成剩余题目包
"""
import os
import yaml
import zipfile
from pathlib import Path

# Web 类别题目配置（16-50题）
WEB_CHALLENGES = [
    {"slug": "web-file-upload", "title": "文件上传漏洞", "difficulty": "easy", "tags": ["vuln:file-upload", "kp:bypass"]},
    {"slug": "web-lfi-basic", "title": "本地文件包含", "difficulty": "easy", "tags": ["vuln:lfi", "kp:path-traversal"]},
    {"slug": "web-command-injection", "title": "命令注入", "difficulty": "easy", "tags": ["vuln:command-injection", "kp:shell"]},
    {"slug": "web-xxe-basic", "title": "XXE 外部实体注入", "difficulty": "easy", "tags": ["vuln:xxe", "kp:xml"]},
    {"slug": "web-csrf-basic", "title": "CSRF 跨站请求伪造", "difficulty": "easy", "tags": ["vuln:csrf", "kp:token"]},
    # ... 继续添加到 50 题
]

def create_web_challenge(config):
    """创建 Web 题目包"""
    slug = config['slug']
    base_dir = Path(f'packs/{slug}')
    base_dir.mkdir(parents=True, exist_ok=True)

    # manifest.yml
    manifest = {
        'spec_version': 'challenge-pack-v1',
        'slug': slug,
        'title': config['title'],
        'category': 'web',
        'difficulty': config['difficulty'],
        'tags': config['tags'],
        'description': {'file': 'statement.md'},
        'hints': [],
        'attachments': [],
        'flag': {'mode': 'dynamic'},
        'runtime': {
            'type': 'container',
            'image': {'source': 'dockerfile', 'context_dir': 'docker', 'dockerfile': 'Dockerfile'},
            'expose': [{'container_port': 80, 'protocol': 'tcp'}]
        },
        'instance': {'ttl': '2h', 'max_per_user': 1},
        'resources': {'cpu': 0.3, 'memory': '128m', 'pids': 50}
    }

    with open(base_dir / 'manifest.yml', 'w', encoding='utf-8') as f:
        yaml.dump(manifest, f, allow_unicode=True, default_flow_style=False)

    # statement.md
    statement = f"""# {config['title']}

## 题目描述

{config.get('description', '待补充题目描述')}

## 目标

找到漏洞并获取 flag。

## 知识点

- {', '.join(config['tags'])}
"""
    with open(base_dir / 'statement.md', 'w', encoding='utf-8') as f:
        f.write(statement)

    # Dockerfile
    docker_dir = base_dir / 'docker'
    docker_dir.mkdir(exist_ok=True)

    dockerfile = """FROM php:7.4-apache
COPY src/ /var/www/html/
RUN chown -R www-data:www-data /var/www/html
EXPOSE 80
CMD ["apache2-foreground"]
"""
    with open(docker_dir / 'Dockerfile', 'w') as f:
        f.write(dockerfile)

    # 源代码
    src_dir = docker_dir / 'src'
    src_dir.mkdir(exist_ok=True)

    index_php = """<?php
echo 'Flag: ' . (getenv('FLAG') ?: 'flag{placeholder}');
?>
"""
    with open(src_dir / 'index.php', 'w') as f:
        f.write(index_php)

    # 打包
    zip_path = f'{slug}.zip'
    with zipfile.ZipFile(zip_path, 'w', zipfile.ZIP_DEFLATED) as zipf:
        for file in base_dir.rglob('*'):
            if file.is_file():
                zipf.write(file, file.relative_to(base_dir.parent))

    print(f"✓ {slug}")
    return zip_path

if __name__ == '__main__':
    print("批量生成 Web 题目...")
    for config in WEB_CHALLENGES[:5]:  # 先生成 5 个测试
        create_web_challenge(config)
    print("完成！")
