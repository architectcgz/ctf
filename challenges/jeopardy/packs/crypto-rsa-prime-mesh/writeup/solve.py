#!/usr/bin/env python3
from __future__ import annotations

import hashlib
import json
import math
from pathlib import Path


def egcd(a: int, b: int) -> tuple[int, int, int]:
    if b == 0:
        return a, 1, 0
    g, x1, y1 = egcd(b, a % b)
    return g, y1, x1 - (a // b) * y1


def modinv(value: int, modulus: int) -> int:
    g, x, _ = egcd(value, modulus)
    if g != 1:
        raise ValueError("inverse does not exist")
    return x % modulus


def xor_bytes(left: bytes, right: bytes) -> bytes:
    return bytes(a ^ b for a, b in zip(left, right))


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = json.loads((root / "attachments" / "mesh.json").read_text(encoding="utf-8"))
    nodes = data["nodes"]
    by_name = {node["name"]: node for node in nodes}

    shared: dict[str, list[tuple[str, int]]] = {node["name"]: [] for node in nodes}
    for idx, left in enumerate(nodes):
        n_left = int(left["n"])
        for right in nodes[idx + 1 :]:
            n_right = int(right["n"])
            g = math.gcd(n_left, n_right)
            if 1 < g < n_left and 1 < g < n_right:
                shared[left["name"]].append((right["name"], g))
                shared[right["name"]].append((left["name"], g))

    fragments: dict[str, str] = {}
    next_hop: dict[str, str] = {}
    for node in nodes:
        name = node["name"]
        edges = shared[name]
        if len(edges) != 2:
            raise SystemExit(f"unexpected degree for {name}: {edges}")

        factors = [prime for _, prime in edges]
        p, q = factors
        n = int(node["n"])
        e = int(node["e"])
        phi = (p - 1) * (q - 1)
        d = modinv(e, phi)

        length = int(node["length"])
        plain = pow(int(node["c"]), d, n).to_bytes(length, "big")

        chosen_neighbor = None
        chosen_prime = None
        for neighbor, prime in edges:
            digest = hashlib.sha1(str(prime).encode("utf-8")).hexdigest()[:8]
            if digest == node["next_hash"]:
                chosen_neighbor = neighbor
                chosen_prime = prime
                break
        if chosen_neighbor is None or chosen_prime is None:
            raise SystemExit(f"next hop not found for {name}")

        mask = hashlib.sha256(str(chosen_prime).encode("utf-8")).digest()[:length]
        fragments[name] = xor_bytes(plain, mask).decode("utf-8")
        next_hop[name] = chosen_neighbor

    order = []
    seen = set()
    current = data["start"]
    while current not in seen:
        seen.add(current)
        order.append(current)
        current = next_hop[current]

    if len(order) != len(nodes):
        raise SystemExit(f"incomplete traversal: {order}")

    flag = "".join(fragments[name] for name in order)
    print(flag)


if __name__ == "__main__":
    main()
