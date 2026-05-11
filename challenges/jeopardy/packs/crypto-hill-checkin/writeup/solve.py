#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def modinv(v: int, m: int) -> int:
    t, new_t = 0, 1
    r, new_r = m, v
    while new_r:
        q = r // new_r
        t, new_t = new_t, t - q * new_t
        r, new_r = new_r, r - q * new_r
    if r != 1:
        raise ValueError("not invertible")
    return t % m


def text_to_pairs(text: str):
    vals = [ord(ch) - ord("A") for ch in text]
    return [(vals[idx], vals[idx + 1]) for idx in range(0, len(vals), 2)]


def matrix_from_blocks(blocks):
    return (blocks[0][0], blocks[1][0], blocks[0][1], blocks[1][1])


def mat_mul(left, right):
    a, b, c, d = left
    e, f, g, h = right
    return (
        (a * e + b * g) % 26,
        (a * f + b * h) % 26,
        (c * e + d * g) % 26,
        (c * f + d * h) % 26,
    )


def mat_inv(mat):
    a, b, c, d = mat
    det = (a * d - b * c) % 26
    inv_det = modinv(det, 26)
    return (
        (d * inv_det) % 26,
        (-b * inv_det) % 26,
        (-c * inv_det) % 26,
        (a * inv_det) % 26,
    )


def decrypt(text: str, inv):
    a, b, c, d = inv
    out = []
    for x0, x1 in text_to_pairs(text):
        out.append(chr(((a * x0 + b * x1) % 26) + ord("A")))
        out.append(chr(((c * x0 + d * x1) % 26) + ord("A")))
    return "".join(out)


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = {}
    for line in (root / "attachments" / "hill.txt").read_text(encoding="utf-8").splitlines():
        key, value = line.split("=", 1)
        values[key] = value
    plain_blocks = text_to_pairs(values["known_plain"])
    cipher_blocks = text_to_pairs(values["known_cipher"])
    plain_mat = matrix_from_blocks(plain_blocks)
    cipher_mat = matrix_from_blocks(cipher_blocks)
    key_mat = mat_mul(cipher_mat, mat_inv(plain_mat))
    body = decrypt(values["target_cipher"], mat_inv(key_mat)).rstrip("X")
    print(f"flag{{{body.lower()}}}")


if __name__ == "__main__":
    main()
