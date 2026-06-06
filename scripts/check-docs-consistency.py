#!/usr/bin/env python3
from pathlib import Path
import subprocess
import sys


ROOT = Path(__file__).resolve().parent.parent
SCRIPT = ROOT / "harness/checks/check_docs_consistency.py"
result = subprocess.run([sys.executable, str(SCRIPT), *sys.argv[1:]], cwd=ROOT, check=False)
sys.exit(result.returncode)
