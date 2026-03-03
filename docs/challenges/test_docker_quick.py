#!/usr/bin/env python3
import subprocess
from pathlib import Path

samples = ['web-sqli-login-bypass', 'misc-python-jail']
for slug in samples:
    dockerfile = Path(f'packs/{slug}/docker/Dockerfile')
    if dockerfile.exists():
        print(f"测试 {slug}...")
        result = subprocess.run(['docker', 'build', '-t', f'test-{slug}', str(dockerfile.parent)],
                              capture_output=True, text=True, timeout=180)
        if result.returncode == 0:
            print(f"  ✅ 构建成功")
            subprocess.run(['docker', 'rmi', f'test-{slug}'], capture_output=True)
        else:
            print(f"  ❌ 构建失败")
            print(result.stderr[:300])
