#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


EXPECTED = {
    "a": 0.08167, "b": 0.01492, "c": 0.02782, "d": 0.04253, "e": 0.12702,
    "f": 0.02228, "g": 0.02015, "h": 0.06094, "i": 0.06966, "j": 0.00153,
    "k": 0.00772, "l": 0.04025, "m": 0.02406, "n": 0.06749, "o": 0.07507,
    "p": 0.01929, "q": 0.00095, "r": 0.05987, "s": 0.06327, "t": 0.09056,
    "u": 0.02758, "v": 0.00978, "w": 0.0236, "x": 0.0015, "y": 0.01974, "z": 0.00074,
}


def decode_column(column: str) -> int:
    best_shift = 0
    best_score = float("inf")
    for shift in range(26):
        decoded = [chr(((ord(ch) - ord("a") - shift) % 26) + ord("a")) for ch in column]
        total = len(decoded)
        counts = {letter: 0 for letter in EXPECTED}
        for ch in decoded:
            counts[ch] += 1
        chi2 = 0.0
        for letter, expected in EXPECTED.items():
            observed = counts[letter]
            expect = total * expected
            chi2 += ((observed - expect) ** 2) / max(expect, 1e-9)
        if chi2 < best_score:
            best_score = chi2
            best_shift = shift
    return best_shift


def decode(text: str, shifts: list[int]) -> str:
    out = []
    idx = 0
    for ch in text:
        if "a" <= ch <= "z":
            shift = shifts[idx % len(shifts)]
            out.append(chr(((ord(ch) - ord("a") - shift) % 26) + ord("a")))
            idx += 1
        else:
            out.append(ch)
    return "".join(out)


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    cipher = (root / "attachments" / "courier.txt").read_text(encoding="utf-8")
    letters = "".join(ch for ch in cipher if "a" <= ch <= "z")
    for key_len in range(2, 9):
        shifts = [decode_column(letters[offset::key_len]) for offset in range(key_len)]
        plain = decode(cipher, shifts)
        match = re.search(r"flag\{[a-z0-9_\-]+\}", plain)
        if match:
            print(match.group(0))
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
