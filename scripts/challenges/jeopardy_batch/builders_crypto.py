from __future__ import annotations

from .helpers import *
from .models import BuildResult, Target

def build_crypto_repeating_key_xor(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    key = b"ledger"
    message = (
        "team ledger note. archive the following recovery sentence carefully. "
        f"only the night shift should know that {flag} belongs to the sealed ticket."
    )
    cipher = repeating_xor(message.encode("utf-8"), key)
    statement = plain_statement(
        intro="一份运维账本被重复密钥 XOR 处理后保存成了二进制样本。",
        steps=[
            "下载附件并判断它不是单字节异或。",
            "恢复重复密钥。",
            "还原明文并提交其中的 flag。",
        ],
        attachments=["ledger.bin"],
        notes=[
            "密钥长度不超过 12。",
            "明文是正常英文句子，不需要爆破 flag 格式。",
        ],
    )
    solution = """# 解法

附件是典型的 repeating-key XOR。先按候选 keysize 计算归一化汉明距离，选出最可能的密钥长度，再把密文按列拆成若干单字节异或问题，逐列按英文频率打分恢复 key。拿到 key 后整体异或即可恢复明文，从中直接取出 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def hamming(left: bytes, right: bytes) -> int:
    return sum((a ^ b).bit_count() for a, b in zip(left, right))


def score(data: bytes) -> float:
    freq = {
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
    total = 0.0
    for byte in data.lower():
        if 32 <= byte <= 126:
            total += freq.get(byte, 0.25)
        else:
            total -= 12.0
    return total


def recover_keysizes(data: bytes) -> list[int]:
    scored = []
    for keysize in range(2, 13):
        blocks = [data[idx : idx + keysize] for idx in range(0, keysize * 4, keysize)]
        if len(blocks[-1]) != keysize:
            continue
        pairs = []
        for idx in range(len(blocks) - 1):
            pairs.append(hamming(blocks[idx], blocks[idx + 1]) / keysize)
        scored.append((sum(pairs) / len(pairs), keysize))
    scored.sort()
    return [keysize for _, keysize in scored]


def recover_key(data: bytes, keysize: int) -> bytes:
    key = bytearray()
    for offset in range(keysize):
        column = data[offset::keysize]
        best = max(
            ((score(bytes(byte ^ candidate for byte in column)), candidate) for candidate in range(256)),
            key=lambda item: item[0],
        )
        key.append(best[1])
    return bytes(key)


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    cipher = (root / "attachments" / "ledger.bin").read_bytes()
    for keysize in recover_keysizes(cipher):
        key = recover_key(cipher, keysize)
        plain = bytes(byte ^ key[idx % len(key)] for idx, byte in enumerate(cipher))
        match = re.search(rb"flag\\{[a-z0-9_\\-]+\\}", plain)
        if match:
            print(match.group().decode())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先比较不同 keysize 的归一化汉明距离。", "拆列后每一列都可以看成单字节异或。"],
        files={"attachments/ledger.bin": cipher},
        attachments=[("attachments/ledger.bin", "ledger.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_crypto_affine_cipher(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    plain = f"internal badge memo says keep {flag} ready for morning dispatch."
    a, b = 5, 8
    cipher = affine_encrypt(plain, a, b)
    statement = plain_statement(
        intro="一张门牌说明被 Affine cipher 处理后留下了可读标点和空格。",
        steps=["识别仿射替换模型。", "恢复参数并解密明文。", "提取其中的 flag。"],
        attachments=["badge.txt"],
        notes=["字母表按小写英文字母处理。"],
    )
    solution = """# 解法

对 Affine cipher 来说，密文满足 `E(x)=a*x+b mod 26`。枚举所有与 26 互素的 `a` 和 0 到 25 的 `b`，逆向解密后按 `flag{...}` 正则或英文可读性筛选即可找到正确明文。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def egcd(a: int, b: int):
    if b == 0:
        return a, 1, 0
    g, x1, y1 = egcd(b, a % b)
    return g, y1, x1 - (a // b) * y1


def modinv(a: int, m: int) -> int:
    g, x, _ = egcd(a, m)
    if g != 1:
        raise ValueError
    return x % m


def decode(text: str, a: int, b: int) -> str:
    inv_a = modinv(a, 26)
    out = []
    for ch in text:
        if "a" <= ch <= "z":
            value = ord(ch) - ord("a")
            out.append(chr(((inv_a * (value - b)) % 26) + ord("a")))
        else:
            out.append(ch)
    return "".join(out)


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    cipher = (root / "attachments" / "badge.txt").read_text(encoding="utf-8")
    for a in range(1, 26):
        if a % 2 == 0 or a == 13:
            continue
        for b in range(26):
            plain = decode(cipher, a, b)
            match = re.search(r"flag\\{[a-z0-9_\\-]+\\}", plain)
            if match:
                print(match.group(0))
                return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先确认字母替换满足线性函数。", "26 下只有与 26 互素的 a 才可逆。"],
        files={"attachments/badge.txt": cipher.encode("utf-8")},
        attachments=[("attachments/badge.txt", "badge.txt")],
        flag_type="static",
        flag_value=flag,
    )

def build_crypto_vigenere_cipher(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    key = "parcel"
    plain = (
        "night courier memo. "
        f"the sealed phrase is {flag} and the route rotates weekly. "
        "confirm the parcel at east gate and keep the ledger hidden until morning shift."
    )
    cipher = vigenere_encrypt(plain, key)
    statement = plain_statement(
        intro="一段投递单摘要使用了多表替换，但保留了空格和标点。",
        steps=["估计密钥长度。", "分列恢复维吉尼亚密钥。", "解密并提取 flag。"],
        attachments=["courier.txt"],
        notes=["密钥长度不超过 8。"],
    )
    solution = """# 解法

先用重复片段或平均重合指数估计 key length，再把每一列视作恺撒移位问题，通过英文频率做 shift 打分。把每列 shift 组合成 key 后整体解密即可。
"""
    solve_py = """#!/usr/bin/env python3
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
        match = re.search(r"flag\\{[a-z0-9_\\-]+\\}", plain)
        if match:
            print(match.group(0))
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先去掉空格和标点，只看字母列。", "每一列都可以按恺撒移位来打分。"],
        files={"attachments/courier.txt": cipher.encode("utf-8")},
        attachments=[("attachments/courier.txt", "courier.txt")],
        flag_type="static",
        flag_value=flag,
    )

def build_crypto_columnar_transposition(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    width = 6
    plain = f"archive route note: submit {flag} before dawn."
    cipher = columnar_encrypt(plain, width)
    statement = plain_statement(
        intro="一份归档便签做了简单列换位，没有再做额外替换。",
        steps=["判断它属于字符重排而不是字母替换。", "枚举合理列宽并重建矩阵。", "恢复明文中的 flag。"],
        attachments=["archive.txt"],
        notes=["列宽不超过 8。"],
    )
    solution = """# 解法

本题是标准列换位。枚举列宽后按总长度和余数还原每一列的长度，把密文切成列再按行读回去。读出的候选明文里出现 `flag{...}` 的就是正确解。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def decrypt(cipher: str, width: int) -> str:
    rows = len(cipher) // width
    extra = len(cipher) % width
    lengths = [rows + (1 if idx < extra else 0) for idx in range(width)]
    cols = []
    cursor = 0
    for length in lengths:
        cols.append(cipher[cursor : cursor + length])
        cursor += length
    out = []
    for row in range(max(lengths)):
        for col in range(width):
            if row < len(cols[col]):
                out.append(cols[col][row])
    return "".join(out)


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    cipher = (root / "attachments" / "archive.txt").read_text(encoding="utf-8")
    for width in range(2, 9):
        plain = decrypt(cipher, width)
        match = re.search(r"flag\\{[a-z0-9_\\-]+\\}", plain)
        if match:
            print(match.group(0))
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["密文长度和列宽一起决定了每列的长度。", "按行写入、按列读出时，解密要反过来。"],
        files={"attachments/archive.txt": cipher.encode("utf-8")},
        attachments=[("attachments/archive.txt", "archive.txt")],
        flag_type="static",
        flag_value=flag,
    )

def build_crypto_hill_cipher(target: Target) -> BuildResult:
    flag = "flag{hillcheckincode}"
    body = "HILLCHECKINCODE"
    matrix = (3, 3, 2, 5)
    known_plain = "HELP"
    known_cipher = hill_encrypt(known_plain, matrix)
    target_cipher = hill_encrypt(body, matrix)
    attachment = "\n".join(
        [
            f"known_plain={known_plain}",
            f"known_cipher={known_cipher}",
            f"target_cipher={target_cipher}",
        ]
    ) + "\n"
    statement = plain_statement(
        intro="一张签到纸给了一段已知明文/密文对，以及另一段待恢复的密文。",
        steps=["根据已知明文块恢复 2x2 Hill 矩阵。", "对目标密文做矩阵逆运算。", "把恢复出的短语按题面规则包装成 flag。"],
        attachments=["hill.txt"],
        notes=["只处理大写英文字母。"],
    )
    solution = """# 解法

已知 `P` 和 `C` 的两个 2 字符块后，可以求出 `K = C * P^{-1} mod 26`。拿到矩阵后再算 `K^{-1}` 去解密目标密文，恢复出的短语转小写后按题面包成 flag 即可。
"""
    solve_py = """#!/usr/bin/env python3
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
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先把已知明文/密文拆成两个 2 字符块。", "解密要用矩阵逆，而不是直接再乘一次。"],
        files={"attachments/hill.txt": attachment.encode("utf-8")},
        attachments=[("attachments/hill.txt", "hill.txt")],
        flag_type="static",
        flag_value=flag,
    )

def build_crypto_lcg_keystream(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    modulus = 2147483647
    multiplier = 48271
    increment = 17389
    state = 1357911
    samples = []
    for _ in range(6):
        state = (multiplier * state + increment) % modulus
        samples.append(state)
    plain = f"telemetry archive says submit {flag} before rotating the night token.".encode("utf-8")
    stream = []
    for _ in range(len(plain)):
        state = (multiplier * state + increment) % modulus
        stream.append(state & 0xFF)
    cipher = bytes(p ^ s for p, s in zip(plain, stream))
    attachment = "samples:\n" + "\n".join(str(item) for item in samples) + "\n\ncipher_hex:\n" + cipher.hex() + "\n"
    statement = plain_statement(
        intro="一份遥测流记录给出了同一发生器的连续输出，以及后续字节流加密得到的密文。",
        steps=["根据输出样本恢复 LCG 参数。", "继续生成后续字节流。", "解开密文并提交 flag。"],
        attachments=["telemetry.txt"],
        notes=["模数已知为 2147483647。"],
    )
    solution = """# 解法

已知模数时，三段连续输出足以解出 `a` 和 `c`。有了完整 LCG 参数后继续往后推，取每个状态的低字节做异或就能解出后续密文。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


MOD = 2147483647


def egcd(a: int, b: int):
    if b == 0:
        return a, 1, 0
    g, x1, y1 = egcd(b, a % b)
    return g, y1, x1 - (a // b) * y1


def modinv(a: int, m: int) -> int:
    g, x, _ = egcd(a, m)
    if g != 1:
        raise ValueError
    return x % m


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "telemetry.txt").read_text(encoding="utf-8")
    sample_part, cipher_part = text.split("cipher_hex:\\n", 1)
    samples = [int(line) for line in sample_part.splitlines() if line.isdigit()]
    cipher = bytes.fromhex(cipher_part.strip())
    a = ((samples[2] - samples[1]) * modinv((samples[1] - samples[0]) % MOD, MOD)) % MOD
    c = (samples[1] - a * samples[0]) % MOD
    state = samples[-1]
    stream = []
    for _ in range(len(cipher)):
        state = (a * state + c) % MOD
        stream.append(state & 0xFF)
    plain = bytes(byte ^ key for byte, key in zip(cipher, stream))
    match = re.search(rb"flag\\{[a-z0-9_\\-]+\\}", plain)
    if not match:
        raise SystemExit("flag not found")
    print(match.group().decode())


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["三段连续输出足够恢复 a 和 c。", "密文字节对应的是后续状态的低字节。"],
        files={"attachments/telemetry.txt": attachment.encode("utf-8")},
        attachments=[("attachments/telemetry.txt", "telemetry.txt")],
        flag_type="static",
        flag_value=flag,
    )

def build_crypto_mt19937_token(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    mt = MiniMT19937(20260509)
    outputs = [mt.rand_u32() for _ in range(624)]
    plain = f"reset token note says {flag} and only the final prediction reveals it.".encode("utf-8")
    stream = bytearray()
    while len(stream) < len(plain):
        stream.extend(struct.pack(">I", mt.rand_u32()))
    cipher = bytes(byte ^ stream[idx] for idx, byte in enumerate(plain))
    attachment = "outputs:\n" + "\n".join(str(item) for item in outputs) + "\n\ncipher_hex:\n" + cipher.hex() + "\n"
    statement = plain_statement(
        intro="一批 32 位伪随机输出被记录了下来，后面同一 PRNG 又被拿去保护了一段重置说明。",
        steps=["根据给出的 624 个输出恢复 MT19937 内部状态。", "预测后续输出。", "解开密文并提交 flag。"],
        attachments=["reset.txt"],
        notes=["输出来自标准 MT19937。"],
    )
    solution = """# 解法

624 个连续输出足以完整克隆 MT19937。先对每个输出做 untemper，恢复状态数组，再按照标准 twist/temper 继续生成字节流，对后续密文异或即可。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
import struct
from pathlib import Path


def unshift_right_xor(value: int, shift: int) -> int:
    result = 0
    for bit in range(32):
        src = 31 - bit
        shifted = src + shift
        known = ((result >> shifted) & 1) if shifted < 32 else 0
        current = ((value >> src) & 1) ^ known
        result |= current << src
    return result


def unshift_left_xor_mask(value: int, shift: int, mask: int) -> int:
    result = 0
    for bit in range(32):
        shifted = bit - shift
        known = ((result >> shifted) & 1) if shifted >= 0 else 0
        mask_bit = (mask >> bit) & 1
        current = ((value >> bit) & 1) ^ (known & mask_bit)
        result |= current << bit
    return result


def untemper(value: int) -> int:
    value = unshift_right_xor(value, 18)
    value = unshift_left_xor_mask(value, 15, 0xEFC60000)
    value = unshift_left_xor_mask(value, 7, 0x9D2C5680)
    value = unshift_right_xor(value, 11)
    return value & 0xFFFFFFFF


class MiniMT:
    def __init__(self, state):
        self.state = list(state)
        self.index = 624

    def twist(self):
        for idx in range(624):
            y = (self.state[idx] & 0x80000000) + (self.state[(idx + 1) % 624] & 0x7FFFFFFF)
            self.state[idx] = self.state[(idx + 397) % 624] ^ (y >> 1)
            if y & 1:
                self.state[idx] ^= 0x9908B0DF
        self.index = 0

    def rand_u32(self):
        if self.index >= 624:
            self.twist()
        y = self.state[self.index]
        self.index += 1
        y ^= y >> 11
        y ^= (y << 7) & 0x9D2C5680
        y ^= (y << 15) & 0xEFC60000
        y ^= y >> 18
        return y & 0xFFFFFFFF


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "reset.txt").read_text(encoding="utf-8")
    output_part, cipher_part = text.split("cipher_hex:\\n", 1)
    outputs = [int(line) for line in output_part.splitlines() if line.isdigit()]
    cipher = bytes.fromhex(cipher_part.strip())
    state = [untemper(value) for value in outputs[:624]]
    mt = MiniMT(state)
    stream = bytearray()
    while len(stream) < len(cipher):
        stream.extend(struct.pack(">I", mt.rand_u32()))
    plain = bytes(byte ^ stream[idx] for idx, byte in enumerate(cipher))
    match = re.search(rb"flag\\{[a-z0-9_\\-]+\\}", plain)
    if not match:
        raise SystemExit("flag not found")
    print(match.group().decode())


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["624 个连续输出足够克隆标准 MT19937。", "恢复状态后还要继续生成后续输出。"],
        files={"attachments/reset.txt": attachment.encode("utf-8")},
        attachments=[("attachments/reset.txt", "reset.txt")],
        flag_type="static",
        flag_value=flag,
    )

def build_crypto_rsa_fermat(target: Target) -> BuildResult:
    body = "closefile"
    flag = f"flag{{{body}}}"
    p = 1000000000039
    q = 1000000000091
    n = p * q
    e = 65537
    m = int.from_bytes(body.encode("utf-8"), "big")
    c = pow(m, e, n)
    attachment = f"n={n}\ne={e}\nc={c}\n"
    statement = plain_statement(
        intro="这份审批单只给了标准 RSA 参数，但模数的两个素因子非常接近。",
        steps=["对模数做 Fermat 分解。", "恢复私钥并解密密文。", "提交得到的 flag。"],
        attachments=["rsa.txt"],
    )
    solution = """# 解法

当 `p` 和 `q` 非常接近时，可以把 `n` 写成 `a^2-b^2` 并快速找到因子。分解后按普通 RSA 流程求 `d`，再解密 `c` 即可。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

from math import isqrt
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = {}
    for line in (root / "attachments" / "rsa.txt").read_text(encoding="utf-8").splitlines():
        key, value = line.split("=", 1)
        values[key] = int(value)
    n = values["n"]
    e = values["e"]
    c = values["c"]
    a = isqrt(n)
    if a * a < n:
        a += 1
    while True:
        b2 = a * a - n
        b = isqrt(b2)
        if b * b == b2:
            p = a - b
            q = a + b
            break
        a += 1
    phi = (p - 1) * (q - 1)
    d = pow(e, -1, phi)
    m = pow(c, d, n)
    data = m.to_bytes((m.bit_length() + 7) // 8, "big")
    print(f"flag{{{data.decode()}}}")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["模数的两个因子只差几十。", "这是 Fermat 分解的典型使用场景。"],
        files={"attachments/rsa.txt": attachment.encode("utf-8")},
        attachments=[("attachments/rsa.txt", "rsa.txt")],
        flag_type="static",
        flag_value=flag,
    )

def build_crypto_rsa_common_modulus(target: Target) -> BuildResult:
    body = "dualroute"
    flag = f"flag{{{body}}}"
    p = 1000000010039
    q = 1000000020049
    n = p * q
    e1 = 65537
    e2 = 17
    m = int.from_bytes(body.encode("utf-8"), "big")
    c1 = pow(m, e1, n)
    c2 = pow(m, e2, n)
    attachment = f"n={n}\ne1={e1}\ne2={e2}\nc1={c1}\nc2={c2}\n"
    statement = plain_statement(
        intro="同一份明文被放在了共模 RSA 系统里，使用了两个互素公钥指数。",
        steps=["识别这是共模条件。", "利用扩展欧几里得合并两份密文。", "恢复明文 flag。"],
        attachments=["common.txt"],
    )
    solution = """# 解法

若 `gcd(e1,e2)=1`，就能找到 `s,t` 使得 `s*e1+t*e2=1`。因此 `m = c1^s * c2^t mod n`，当系数为负时改用模逆即可。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def egcd(a: int, b: int):
    if b == 0:
        return a, 1, 0
    g, x1, y1 = egcd(b, a % b)
    return g, y1, x1 - (a // b) * y1


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = {}
    for line in (root / "attachments" / "common.txt").read_text(encoding="utf-8").splitlines():
        key, value = line.split("=", 1)
        values[key] = int(value)
    n = values["n"]
    e1 = values["e1"]
    e2 = values["e2"]
    c1 = values["c1"]
    c2 = values["c2"]
    _, s, t = egcd(e1, e2)
    if s < 0:
        c1 = pow(c1, -1, n)
        s = -s
    if t < 0:
        c2 = pow(c2, -1, n)
        t = -t
    m = (pow(c1, s, n) * pow(c2, t, n)) % n
    data = m.to_bytes((m.bit_length() + 7) // 8, "big")
    print(f"flag{{{data.decode()}}}")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先检查两个指数是否互素。", "负指数要换成模逆。"],
        files={"attachments/common.txt": attachment.encode("utf-8")},
        attachments=[("attachments/common.txt", "common.txt")],
        flag_type="static",
        flag_value=flag,
    )

def build_crypto_rsa_broadcast(target: Target) -> BuildResult:
    body = "tricastok"
    flag = f"flag{{{body}}}"
    e = 3
    m = int.from_bytes(body.encode("utf-8"), "big")
    moduli = [
        1000000000039 * 1000000000091,
        1000000010039 * 1000000020049,
        1000000030067 * 1000000040041,
    ]
    ciphers = [pow(m, e, n) for n in moduli]
    attachment = "\n".join(
        ["e=3", *(f"n{idx + 1}={value}" for idx, value in enumerate(moduli)), *(f"c{idx + 1}={value}" for idx, value in enumerate(ciphers))]
    ) + "\n"
    statement = plain_statement(
        intro="同一份通知被发给了三个接收端，公钥指数相同且都用了低指数 e=3。",
        steps=["对三组模数和密文做 CRT 合并。", "对合并结果取整数立方根。", "还原并提交 flag。"],
        attachments=["broadcast.txt"],
    )
    solution = """# 解法

这是低指数广播攻击。把三份同明文密文通过 CRT 合并得到 `m^3`，再做整数立方根即可恢复原文。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def crt(ns, cs):
    total = 0
    big_n = 1
    for n in ns:
        big_n *= n
    for n, c in zip(ns, cs):
        m = big_n // n
        total += c * m * pow(m, -1, n)
    return total % big_n


def icbrt(value: int) -> int:
    low, high = 0, value
    while low <= high:
        mid = (low + high) // 2
        cube = mid * mid * mid
        if cube == value:
            return mid
        if cube < value:
            low = mid + 1
        else:
            high = mid - 1
    return high


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = {}
    for line in (root / "attachments" / "broadcast.txt").read_text(encoding="utf-8").splitlines():
        key, value = line.split("=", 1)
        values[key] = int(value)
    ns = [values["n1"], values["n2"], values["n3"]]
    cs = [values["c1"], values["c2"], values["c3"]]
    cube = crt(ns, cs)
    m = icbrt(cube)
    data = m.to_bytes((m.bit_length() + 7) // 8, "big")
    print(f"flag{{{data.decode()}}}")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["同一明文、同一低指数、不同模数。", "先 CRT，再开整数立方根。"],
        files={"attachments/broadcast.txt": attachment.encode("utf-8")},
        attachments=[("attachments/broadcast.txt", "broadcast.txt")],
        flag_type="static",
        flag_value=flag,
    )

def build_crypto_dsa_nonce_reuse(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    p = 30803
    q = 1531
    g = pow(2, (p - 1) // q, p)
    x = 733
    y = pow(g, x, p)
    k = 902
    h1 = 871
    h2 = 1204
    r = pow(g, k, p) % q
    s1 = (modinv(k, q) * (h1 + x * r)) % q
    s2 = (modinv(k, q) * (h2 + x * r)) % q
    key = zlib.crc32(str(x).encode("utf-8")).to_bytes(4, "big")
    cipher = repeating_xor(flag.encode("utf-8"), key)
    attachment = "\n".join(
        [
            f"p={p}",
            f"q={q}",
            f"g={g}",
            f"y={y}",
            f"h1={h1}",
            f"h2={h2}",
            f"r={r}",
            f"s1={s1}",
            f"s2={s2}",
            f"cipher_hex={cipher.hex()}",
        ]
    ) + "\n"
    statement = plain_statement(
        intro="同一个 DSA 随机数被复用了两次，私钥恢复后还要再解开一段短密文。",
        steps=["利用两次签名恢复随机数 k。", "再恢复私钥 x。", "按题面规则导出密钥并解开 flag。"],
        attachments=["signatures.txt"],
    )
    solution = """# 解法

随机数复用时有：

`k = (h1 - h2) / (s1 - s2) mod q`

再代回：

`x = (s1 * k - h1) / r mod q`

本题用 `crc32(str(x))` 作为 4 字节重复异或密钥，所以恢复私钥后再异或一遍即可拿到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import zlib
from pathlib import Path


def egcd(a: int, b: int):
    if b == 0:
        return a, 1, 0
    g, x1, y1 = egcd(b, a % b)
    return g, y1, x1 - (a // b) * y1


def modinv(a: int, m: int) -> int:
    g, x, _ = egcd(a, m)
    if g != 1:
        raise ValueError
    return x % m


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = {}
    for line in (root / "attachments" / "signatures.txt").read_text(encoding="utf-8").splitlines():
        key, value = line.split("=", 1)
        values[key] = value
    q = int(values["q"])
    h1 = int(values["h1"])
    h2 = int(values["h2"])
    r = int(values["r"])
    s1 = int(values["s1"])
    s2 = int(values["s2"])
    cipher = bytes.fromhex(values["cipher_hex"])
    k = ((h1 - h2) * modinv((s1 - s2) % q, q)) % q
    x = ((s1 * k - h1) * modinv(r, q)) % q
    key = zlib.crc32(str(x).encode("utf-8")).to_bytes(4, "big")
    plain = bytes(byte ^ key[idx % len(key)] for idx, byte in enumerate(cipher))
    print(plain.decode())


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先从两次签名里恢复同一个 k。", "最后一层只是很短的重复密钥异或。"],
        files={"attachments/signatures.txt": attachment.encode("utf-8")},
        attachments=[("attachments/signatures.txt", "signatures.txt")],
        flag_type="static",
        flag_value=flag,
    )
