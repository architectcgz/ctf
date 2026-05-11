#!/usr/bin/env python3
from __future__ import annotations

from collections import deque
from pathlib import Path
import subprocess


MOVES = ["l", "r", "u", "d"]


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    rows = []
    target = 5
    for line in (root / "attachments" / "state_table.txt").read_text(encoding="utf-8").splitlines():
        if line.startswith("row"):
            rows.append([int(item) for item in line.split(":")[1].split()])
    queue = deque([(0, "")])
    seen = {(0, "")}
    while queue:
        state, path = queue.popleft()
        if state == target:
            proc = subprocess.run([binary], input=path + "\n", text=True, capture_output=True)
            if "ok" in proc.stdout:
                print(f"flag{{{path}}}")
                return
        if state >= len(rows):
            continue
        for idx, nxt in enumerate(rows[state]):
            candidate = path + MOVES[idx]
            if len(candidate) > 8:
                continue
            node = (nxt, candidate)
            if node not in seen:
                seen.add(node)
                queue.append((nxt, candidate))
    raise SystemExit("path not found")


if __name__ == "__main__":
    main()
