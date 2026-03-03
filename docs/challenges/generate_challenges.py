#!/usr/bin/env python3
"""
CTF 题目包批量生成脚本
根据题目列表自动生成完整的题目包结构
"""
import os
import yaml
import zipfile
from pathlib import Path

# 题目模板配置
TEMPLATES = {
    'web-static': {
        'dockerfile': '''FROM nginx:alpine
COPY src/ /usr/share/nginx/html/
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]''',
        'index_template': '''<!DOCTYPE html>
<html>
<head><title>{title}</title></head>
<body>
<h1>{title}</h1>
{content}
</body>
</html>'''
    },
    'web-php': {
        'dockerfile': '''FROM php:7.4-apache
COPY src/ /var/www/html/
RUN chown -R www-data:www-data /var/www/html
EXPOSE 80
CMD ["apache2-foreground"]''',
    }
}

def create_challenge_pack(slug, title, category, difficulty, tags, description, challenge_type='web-static'):
    """创建单个题目包"""
    base_dir = Path(f'/home/azhi/workspace/projects/ctf/docs/challenges/packs/{slug}')
    base_dir.mkdir(parents=True, exist_ok=True)

    # 创建 manifest.yml
    manifest = {
        'spec_version': 'challenge-pack-v1',
        'slug': slug,
        'title': title,
        'category': category,
        'difficulty': difficulty,
        'tags': tags,
        'description': {'file': 'statement.md'},
        'hints': [],
        'attachments': [],
        'flag': {'mode': 'dynamic'},
        'runtime': {
            'type': 'container',
            'image': {
                'source': 'dockerfile',
                'context_dir': 'docker',
                'dockerfile': 'Dockerfile'
            },
            'expose': [{'container_port': 80, 'protocol': 'tcp'}]
        },
        'instance': {'ttl': '2h', 'max_per_user': 1},
        'resources': {'cpu': 0.3, 'memory': '128m', 'pids': 50}
    }

    with open(base_dir / 'manifest.yml', 'w', encoding='utf-8') as f:
        yaml.dump(manifest, f, allow_unicode=True, default_flow_style=False)

    # 创建 statement.md
    statement = f'''# {title}

## 题目描述

{description}

## 目标

找到隐藏的 flag 并提交。

## 访问方式

点击"开始挑战"后获取靶机地址。
'''

    with open(base_dir / 'statement.md', 'w', encoding='utf-8') as f:
        f.write(statement)

    # 创建 docker 目录
    docker_dir = base_dir / 'docker'
    docker_dir.mkdir(exist_ok=True)

    # 创建 Dockerfile
    template = TEMPLATES.get(challenge_type, TEMPLATES['web-static'])
    with open(docker_dir / 'Dockerfile', 'w') as f:
        f.write(template['dockerfile'])

    # 创建 src 目录
    src_dir = docker_dir / 'src'
    src_dir.mkdir(exist_ok=True)

    return base_dir

if __name__ == '__main__':
    print("CTF 题目包生成脚本")
    print("使用方法：修改此脚本添加题目配置，然后运行")
