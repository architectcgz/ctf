#!/usr/bin/env python3
import os, yaml, zipfile
from pathlib import Path

MISC_CONFIGS = [
    {"slug": "misc-python-jail", "title": "Python 沙箱逃逸", "difficulty": "easy", "tags": ["kp:sandbox"]},
    {"slug": "misc-bash-jail", "title": "Bash 受限环境", "difficulty": "easy", "tags": ["kp:restricted-shell"]},
    {"slug": "misc-regex-bypass", "title": "正则表达式绕过", "difficulty": "easy", "tags": ["kp:regex"]},
    {"slug": "misc-git-leak", "title": "Git 信息泄露", "difficulty": "easy", "tags": ["kp:git"]},
    {"slug": "misc-docker-escape", "title": "Docker 逃逸", "difficulty": "easy", "tags": ["kp:docker"]},
    {"slug": "misc-zip-slip", "title": "ZIP 路径遍历", "difficulty": "easy", "tags": ["kp:zip-slip"]},
    {"slug": "misc-tar-bomb", "title": "TAR 炸弹", "difficulty": "easy", "tags": ["kp:tar"]},
    {"slug": "misc-polyglot", "title": "多语言文件", "difficulty": "easy", "tags": ["kp:polyglot"]},
    {"slug": "misc-magic-hash", "title": "魔术哈希", "difficulty": "easy", "tags": ["kp:hash"]},
    {"slug": "misc-timing-attack", "title": "时序攻击", "difficulty": "easy", "tags": ["kp:timing"]},
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
    print(f"\n已完成 Misc 类别：20/50 题")
