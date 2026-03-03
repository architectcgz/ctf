#!/usr/bin/env python3
import yaml, subprocess, time
from pathlib import Path

results = {'success': [], 'failed': []}

for manifest_path in sorted(Path('packs').rglob('manifest.yml')):
    slug = manifest_path.parent.name
    with open(manifest_path) as f:
        data = yaml.safe_load(f)

    runtime = data.get('runtime', {})
    if runtime.get('type') == 'container':
        dockerfile = manifest_path.parent / 'docker' / 'Dockerfile'
        if dockerfile.exists():
            print(f"测试 {slug}...", end=' ')
            try:
                result = subprocess.run(
                    ['docker', 'build', '-q', '-t', f'test-{slug}', str(dockerfile.parent)],
                    capture_output=True, timeout=120
                )
                if result.returncode == 0:
                    subprocess.run(['docker', 'rmi', '-f', f'test-{slug}'], capture_output=True)
                    results['success'].append(slug)
                    print('✅')
                else:
                    results['failed'].append((slug, 'build failed'))
                    print('❌')
            except subprocess.TimeoutExpired:
                results['failed'].append((slug, 'timeout'))
                print('⏱️')
            except Exception as e:
                results['failed'].append((slug, str(e)))
                print('❌')
    else:
        results['success'].append(slug)

print(f"\n成功: {len(results['success'])}")
print(f"失败: {len(results['failed'])}")
if results['failed']:
    for slug, err in results['failed'][:10]:
        print(f"  {slug}: {err}")
