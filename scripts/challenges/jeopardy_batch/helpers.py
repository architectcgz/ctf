from __future__ import annotations

import base64
import binascii
import bz2
import gzip
import io
import json
import os
import random
import re
import shutil
import sqlite3
import struct
import subprocess
import tarfile
import tempfile
import textwrap
import zlib
from pathlib import Path
from zipfile import ZIP_DEFLATED, ZipFile

def slug_flag(slug: str) -> str:
    return f"flag{{{slug.replace('-', '_')}}}"

def rng_for(slug: str) -> random.Random:
    seed = int.from_bytes(slug.encode("utf-8"), "little") % (2**32)
    return random.Random(seed)

def to_hex_lines(data: bytes, width: int = 16) -> str:
    chunks = []
    for idx in range(0, len(data), width):
        row = data[idx : idx + width]
        chunks.append(" ".join(f"{byte:02x}" for byte in row))
    return "\n".join(chunks)

def hamming_distance(left: bytes, right: bytes) -> int:
    return sum((a ^ b).bit_count() for a, b in zip(left, right))

def english_score(data: bytes) -> float:
    frequencies = {
        ord(" "): 13.0,
        ord("e"): 12.7,
        ord("t"): 9.1,
        ord("a"): 8.2,
        ord("o"): 7.5,
        ord("i"): 7.0,
        ord("n"): 6.7,
        ord("s"): 6.3,
        ord("h"): 6.1,
        ord("r"): 6.0,
    }
    score = 0.0
    for byte in data.lower():
        if 32 <= byte <= 126:
            score += frequencies.get(byte, 0.3)
        else:
            score -= 12.0
    return score

def compile_c(source: str, output_path: Path, extra_flags: list[str] | None = None) -> None:
    extra_flags = extra_flags or []
    with tempfile.TemporaryDirectory() as tmp_dir:
        src_path = Path(tmp_dir) / "challenge.c"
        src_path.write_text(source, encoding="utf-8")
        cmd = ["gcc", "-O0", "-g", "-o", str(output_path), str(src_path), *extra_flags]
        subprocess.run(cmd, check=True)

def compile_pwn_c(source: str, output_path: Path) -> None:
    compile_c(
        source,
        output_path,
        extra_flags=["-fno-stack-protector", "-no-pie"],
    )

def build_tar_gz(entries: dict[str, bytes], mode_map: dict[str, int] | None = None) -> bytes:
    mode_map = mode_map or {}
    buf = io.BytesIO()
    with tarfile.open(fileobj=buf, mode="w:gz") as tf:
        for name, payload in entries.items():
            info = tarfile.TarInfo(name)
            info.size = len(payload)
            info.mode = mode_map.get(name, 0o644)
            tf.addfile(info, io.BytesIO(payload))
    return buf.getvalue()

def build_zip(entries: dict[str, bytes]) -> bytes:
    buf = io.BytesIO()
    with ZipFile(buf, "w", compression=ZIP_DEFLATED) as zf:
        for name, payload in entries.items():
            zf.writestr(name, payload)
    return buf.getvalue()

def plain_statement(
    intro: str,
    steps: list[str],
    attachments: list[str] | None = None,
    notes: list[str] | None = None,
    access: list[str] | None = None,
) -> str:
    lines = [intro.strip(), "", "## 目标", ""]
    for idx, step in enumerate(steps, start=1):
        lines.append(f"{idx}. {step}")
    if attachments:
        lines.extend(["", "## 附件"])
        for item in attachments:
            lines.append(f"- `{item}`")
    if access:
        lines.extend(["", "## 访问方式"])
        for item in access:
            lines.append(f"- `{item}`")
    if notes:
        lines.extend(["", "## 补充说明"])
        for item in notes:
            lines.append(f"- {item}")
    return "\n".join(lines).rstrip() + "\n"

def sqlite_bytes(sql_commands: list[str], params: list[tuple] | None = None) -> bytes:
    params = params or []
    with tempfile.TemporaryDirectory() as tmp_dir:
        db_path = Path(tmp_dir) / "sample.db"
        conn = sqlite3.connect(db_path)
        cur = conn.cursor()
        for command in sql_commands:
            cur.executescript(command)
        for sql, values in params:
            cur.execute(sql, values)
        conn.commit()
        conn.close()
        return db_path.read_bytes()

def checksum16(data: bytes) -> int:
    if len(data) % 2:
        data += b"\x00"
    total = 0
    for idx in range(0, len(data), 2):
        total += int.from_bytes(data[idx : idx + 2], "big")
    while total >> 16:
        total = (total & 0xFFFF) + (total >> 16)
    return (~total) & 0xFFFF

def build_ipv4_packet(src_ip: bytes, dst_ip: bytes, proto: int, payload: bytes) -> bytes:
    version_ihl = 0x45
    total_length = 20 + len(payload)
    header = struct.pack(
        "!BBHHHBBH4s4s",
        version_ihl,
        0,
        total_length,
        0x1234,
        0x4000,
        64,
        proto,
        0,
        src_ip,
        dst_ip,
    )
    checksum = checksum16(header)
    return struct.pack(
        "!BBHHHBBH4s4s",
        version_ihl,
        0,
        total_length,
        0x1234,
        0x4000,
        64,
        proto,
        checksum,
        src_ip,
        dst_ip,
    ) + payload

def build_tcp_segment(src_ip: bytes, dst_ip: bytes, src_port: int, dst_port: int, payload: bytes) -> bytes:
    header = struct.pack("!HHIIHHHH", src_port, dst_port, 1, 1, 0x5018, 4096, 0, 0)
    pseudo = src_ip + dst_ip + struct.pack("!BBH", 0, 6, len(header) + len(payload))
    checksum = checksum16(pseudo + header + payload)
    return struct.pack("!HHIIHHHH", src_port, dst_port, 1, 1, 0x5018, 4096, checksum, 0) + payload

def build_udp_datagram(src_ip: bytes, dst_ip: bytes, src_port: int, dst_port: int, payload: bytes) -> bytes:
    length = 8 + len(payload)
    header = struct.pack("!HHHH", src_port, dst_port, length, 0)
    pseudo = src_ip + dst_ip + struct.pack("!BBH", 0, 17, length)
    checksum = checksum16(pseudo + header + payload)
    return struct.pack("!HHHH", src_port, dst_port, length, checksum) + payload

def wrap_ethernet(ip_packet: bytes) -> bytes:
    return b"\xaa\xbb\xcc\xdd\xee\xff" + b"\x11\x22\x33\x44\x55\x66" + b"\x08\x00" + ip_packet

def build_pcap(frames: list[bytes]) -> bytes:
    buf = io.BytesIO()
    buf.write(struct.pack("<IHHIIII", 0xA1B2C3D4, 2, 4, 0, 0, 65535, 1))
    ts_sec = 1715251200
    for idx, frame in enumerate(frames):
        buf.write(struct.pack("<IIII", ts_sec + idx, 0, len(frame), len(frame)))
        buf.write(frame)
    return buf.getvalue()

def build_dns_query(name: str) -> bytes:
    labels = name.split(".")
    qname = b"".join(bytes([len(label)]) + label.encode("ascii") for label in labels) + b"\x00"
    header = struct.pack("!HHHHHH", 0x1337, 0x0100, 1, 0, 0, 0)
    question = qname + struct.pack("!HH", 1, 1)
    return header + question

def png_chunk(chunk_type: bytes, data: bytes) -> bytes:
    crc = zlib.crc32(chunk_type + data) & 0xFFFFFFFF
    return struct.pack("!I", len(data)) + chunk_type + data + struct.pack("!I", crc)

def compile_reverse_bundle(source: str) -> tuple[bytes, bytes, bytes]:
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_path = Path(tmp_dir) / "challenge.bin"
        compile_c(source, bin_path)
        asm = subprocess.check_output(["objdump", "-d", "-Mintel", str(bin_path)], text=True).encode("utf-8")
        try:
            rodata = subprocess.check_output(["objdump", "-s", "-j", ".rodata", str(bin_path)], text=True).encode("utf-8")
        except subprocess.CalledProcessError:
            rodata = b""
        return bin_path.read_bytes(), asm, rodata

def egcd(a: int, b: int) -> tuple[int, int, int]:
    if b == 0:
        return a, 1, 0
    g, x1, y1 = egcd(b, a % b)
    return g, y1, x1 - (a // b) * y1

def modinv(value: int, modulus: int) -> int:
    g, x, _ = egcd(value, modulus)
    if g != 1:
        raise ValueError(f"{value} 在模 {modulus} 下不可逆")
    return x % modulus

def repeating_xor(data: bytes, key: bytes) -> bytes:
    return bytes(byte ^ key[idx % len(key)] for idx, byte in enumerate(data))

def affine_encrypt(text: str, a: int, b: int) -> str:
    result = []
    for ch in text:
        if "a" <= ch <= "z":
            value = ord(ch) - ord("a")
            result.append(chr(((a * value + b) % 26) + ord("a")))
        else:
            result.append(ch)
    return "".join(result)

def vigenere_encrypt(text: str, key: str) -> str:
    shifts = [ord(ch) - ord("a") for ch in key]
    result = []
    index = 0
    for ch in text:
        if "a" <= ch <= "z":
            shift = shifts[index % len(shifts)]
            result.append(chr(((ord(ch) - ord("a") + shift) % 26) + ord("a")))
            index += 1
        else:
            result.append(ch)
    return "".join(result)

def columnar_encrypt(text: str, width: int) -> str:
    rows = [text[idx : idx + width] for idx in range(0, len(text), width)]
    output: list[str] = []
    for col in range(width):
        for row in rows:
            if col < len(row):
                output.append(row[col])
    return "".join(output)

def hill_chunk(text: str) -> list[int]:
    values = [ord(ch) - ord("A") for ch in text]
    if len(values) % 2:
        values.append(ord("X") - ord("A"))
    return values

def hill_encrypt(text: str, matrix: tuple[int, int, int, int]) -> str:
    a, b, c, d = matrix
    values = hill_chunk(text)
    out: list[str] = []
    for idx in range(0, len(values), 2):
        x0, x1 = values[idx], values[idx + 1]
        out.append(chr(((a * x0 + b * x1) % 26) + ord("A")))
        out.append(chr(((c * x0 + d * x1) % 26) + ord("A")))
    return "".join(out)

class MiniMT19937:
    def __init__(self, seed: int) -> None:
        self.index = 624
        self.state = [0] * 624
        self.state[0] = seed & 0xFFFFFFFF
        for idx in range(1, 624):
            self.state[idx] = (
                1812433253 * (self.state[idx - 1] ^ (self.state[idx - 1] >> 30)) + idx
            ) & 0xFFFFFFFF

    def twist(self) -> None:
        for idx in range(624):
            y = (self.state[idx] & 0x80000000) + (self.state[(idx + 1) % 624] & 0x7FFFFFFF)
            self.state[idx] = self.state[(idx + 397) % 624] ^ (y >> 1)
            if y & 1:
                self.state[idx] ^= 0x9908B0DF
        self.index = 0

    def rand_u32(self) -> int:
        if self.index >= 624:
            self.twist()
        y = self.state[self.index]
        self.index += 1
        y ^= y >> 11
        y ^= (y << 7) & 0x9D2C5680
        y ^= (y << 15) & 0xEFC60000
        y ^= y >> 18
        return y & 0xFFFFFFFF
