from __future__ import annotations

from .helpers import *
from .models import BuildResult, Target

def build_reverse_native_const_array(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    encoded = [ord(ch) + idx for idx, ch in enumerate(flag)]
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        #include <string.h>

        static const unsigned char encoded[] = {{{", ".join(str(v) for v in encoded)}}};

        int main(void) {{
            char buf[128] = {{0}};
            if (!fgets(buf, sizeof(buf), stdin)) return 1;
            buf[strcspn(buf, "\\n")] = 0;
            if (strlen(buf) != sizeof(encoded)) {{
                puts("nope");
                return 1;
            }}
            for (size_t i = 0; i < sizeof(encoded); ++i) {{
                if (((unsigned char)buf[i] + i) != encoded[i]) {{
                    puts("nope");
                    return 1;
                }}
            }}
            puts("ok");
            return 0;
        }}
        """
    )
    binary, asm, rodata = compile_reverse_bundle(source)
    constants = "\n".join(str(value) for value in encoded).encode("utf-8")
    statement = plain_statement(
        intro="附件包含一份原生二进制和它的 objdump 输出，目标是回推正确输入。",
        steps=["查看校验逻辑。", "逆推出每一位应该是什么字符。", "提交恢复出的 flag。"],
        attachments=["challenge.bin", "challenge.asm", "challenge.rodata.txt", "constants.txt"],
    )
    solution = """# 解法

检查逻辑本质是 `input[i] + i == encoded[i]`。常量表可从 rodata 或题包附带的辅助转储里提出来，对每个位置减去索引即可恢复原始 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = [int(line.strip()) for line in (root / "attachments" / "constants.txt").read_text(encoding="utf-8").splitlines() if line.strip()]
    text = "".join(chr(value - idx) for idx, value in enumerate(values))
    print(text)


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["常量表里每个值都比原字符大一个位置偏移。", "objdump 的 rodata 已经足够解题。"],
        files={
            "attachments/challenge.bin": binary,
            "attachments/challenge.asm": asm,
            "attachments/challenge.rodata.txt": rodata,
            "attachments/constants.txt": constants,
        },
        attachments=[
            ("attachments/challenge.bin", "challenge.bin"),
            ("attachments/challenge.asm", "challenge.asm"),
            ("attachments/challenge.rodata.txt", "challenge.rodata.txt"),
            ("attachments/constants.txt", "constants.txt"),
        ],
        flag_type="static",
        flag_value=flag,
    )

def build_reverse_powershell_xor(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    key = 0x2A
    data = [byte ^ key for byte in flag.encode("utf-8")]
    ps1 = textwrap.dedent(
        f"""\
        $key = {key}
        $data = @({", ".join(str(v) for v in data)})
        $out = ""
        foreach ($b in $data) {{
          $out += [char]($b -bxor $key)
        }}
        if ($out -eq (Read-Host "code")) {{
          "ok"
        }}
        """
    ).encode("utf-8")
    statement = plain_statement(
        intro="附件是一段 PowerShell crackme，核心逻辑是编码数组和一段按位运算。",
        steps=["读懂脚本里的数组和 key。", "逆运算恢复原字符串。", "提交得到的 flag。"],
        attachments=["check.ps1"],
    )
    solution = """# 解法

脚本把每个字节和固定 key 做了 XOR。把数组逐个与 key 逆运算即可恢复原文，不需要跑 Windows 环境。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "check.ps1").read_text(encoding="utf-8")
    key = int(re.search(r"\\$key = (\\d+)", text).group(1))
    numbers = [int(item) for item in re.search(r"@\\(([^)]*)\\)", text, re.S).group(1).split(",")]
    print(bytes(value ^ key for value in numbers).decode())


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["PowerShell 的 `-bxor` 就是按位异或。", "数组里每个数字都对应一个字符。"],
        files={"attachments/check.ps1": ps1},
        attachments=[("attachments/check.ps1", "check.ps1")],
        flag_type="static",
        flag_value=flag,
    )

def build_reverse_batch_substring(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    pool = f"xx{flag}yy"
    parts = []
    cursor = 2
    for chunk_start in range(0, len(flag), 8):
        chunk = flag[chunk_start : chunk_start + 8]
        parts.append(f"%pool:~{cursor},{len(chunk)}%")
        cursor += len(chunk)
    batch = textwrap.dedent(
        f"""\
        @echo off
        set pool={pool}
        set out={''.join(parts)}
        if "%1"=="%out%" echo ok
        """
    ).encode("utf-8")
    statement = plain_statement(
        intro="附件是一段 Windows Batch 脚本，核心在变量池切片拼接。",
        steps=["跟踪 `%var:~start,len%` 语法。", "按顺序拼出完整字符串。", "提交恢复出的 flag。"],
        attachments=["gate.bat"],
    )
    solution = """# 解法

Batch 的 `%pool:~start,len%` 表示字符串切片。按脚本顺序把几段切片拼起来，就能直接恢复目标 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "gate.bat").read_text(encoding="utf-8")
    pool = re.search(r"set pool=(.*)", text).group(1)
    parts = re.findall(r"%pool:~(\\d+),(\\d+)%", text)
    out = "".join(pool[int(start): int(start) + int(length)] for start, length in parts)
    print(out)


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["Batch 切片语法和 shell 不一样。", "每段切片长度都已经给出来了。"],
        files={"attachments/gate.bat": batch},
        attachments=[("attachments/gate.bat", "gate.bat")],
        flag_type="static",
        flag_value=flag,
    )

def build_reverse_native_state_machine(target: Target) -> BuildResult:
    flag = "flag{ldd}"
    solution_path = "ldd"
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        #include <string.h>

        static const int table[4][4] = {{
          {{1,0,2,0}},
          {{1,2,1,3}},
          {{2,2,0,4}},
          {{3,4,3,5}},
        }};

        int main(void) {{
            char buf[64] = {{0}};
            int state = 0;
            if (!fgets(buf, sizeof(buf), stdin)) return 1;
            buf[strcspn(buf, "\\n")] = 0;
            for (size_t i = 0; i < strlen(buf); ++i) {{
                int col = (buf[i] == 'l') ? 0 : (buf[i] == 'r') ? 1 : (buf[i] == 'u') ? 2 : 3;
                state = table[state][col];
            }}
            if (state == 5 && strcmp(buf, "{solution_path}") == 0) {{
                puts("ok");
            }} else {{
                puts("nope");
            }}
            return 0;
        }}
        """
    )
    binary, asm, _ = compile_reverse_bundle(source)
    table_txt = "row0: 1 0 2 0\nrow1: 1 2 1 3\nrow2: 2 2 0 4\nrow3: 3 4 3 5\ntarget_state=5\n"
    statement = plain_statement(
        intro="附件里的原生样本本质上是一个小状态机，需要找出能走到终点的输入串。",
        steps=["读懂状态转移。", "求一条从起点到终点的可达路径。", "把路径按题面包装成 flag。"],
        attachments=["challenge.bin", "challenge.asm", "state_table.txt"],
    )
    solution = """# 解法

状态机题最稳的办法就是把转移表整理出来，再用 BFS / DFS 从起点找终点。本题找到的唯一路径是 `ldd`，提交时按题面写成 `flag{ldd}`。
"""
    solve_py = """#!/usr/bin/env python3
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
            proc = subprocess.run([binary], input=path + "\\n", text=True, capture_output=True)
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
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先把转移表抽出来。", "求路径问题用 BFS 往往最稳。"],
        files={
            "attachments/challenge.bin": binary,
            "attachments/challenge.asm": asm,
            "attachments/state_table.txt": table_txt.encode("utf-8"),
        },
        attachments=[
            ("attachments/challenge.bin", "challenge.bin"),
            ("attachments/challenge.asm", "challenge.asm"),
            ("attachments/state_table.txt", "state_table.txt"),
        ],
        flag_type="static",
        flag_value=flag,
    )

def build_reverse_native_crc32(target: Target) -> BuildResult:
    flag = "flag{HUNT}"
    expected = zlib.crc32(b"HUNT") & 0xFFFFFFFF
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        #include <string.h>
        #include <zlib.h>
        int main(void) {{
            char buf[32] = {{0}};
            if (!fgets(buf, sizeof(buf), stdin)) return 1;
            buf[strcspn(buf, "\\n")] = 0;
            if (strlen(buf) == 4 && crc32(0L, (const unsigned char*)buf, 4) == 0x{expected:08x}) puts("ok");
            else puts("nope");
            return 0;
        }}
        """
    )
    with tempfile.TemporaryDirectory() as tmp_dir:
        src = Path(tmp_dir) / "challenge.c"
        src.write_text(source, encoding="utf-8")
        bin_path = Path(tmp_dir) / "challenge.bin"
        subprocess.run(["gcc", "-O0", "-g", "-o", str(bin_path), str(src), "-lz"], check=True)
        asm = subprocess.check_output(["objdump", "-d", "-Mintel", str(bin_path)], text=True).encode("utf-8")
        binary = bin_path.read_bytes()
    statement = plain_statement(
        intro="附件中的二进制只接受 4 个大写字符输入，并对它们做 CRC32 校验。",
        steps=["定位目标 CRC32 常量。", "恢复满足条件的 4 字符串。", "按题面包装成 flag。"],
        attachments=["challenge.bin", "challenge.asm"],
    )
    solution = """# 解法

本题只需要 4 个大写字符，直接对 `A-Z` 穷举 26^4 也完全可做。找到满足目标 CRC32 的输入 `HUNT` 后，提交 `flag{HUNT}`。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import itertools
import zlib


TARGET = 0x%08x


def main() -> None:
    alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    for chars in itertools.product(alphabet, repeat=4):
        word = "".join(chars)
        if zlib.crc32(word.encode()) & 0xFFFFFFFF == TARGET:
            print(f"flag{{{word}}}")
            return
    raise SystemExit("not found")


if __name__ == "__main__":
    main()
""" % expected
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["输入空间只有 4 位大写字母。", "先把目标 CRC 常量抠出来。"],
        files={"attachments/challenge.bin": binary, "attachments/challenge.asm": asm},
        attachments=[("attachments/challenge.bin", "challenge.bin"), ("attachments/challenge.asm", "challenge.asm")],
        flag_type="static",
        flag_value=flag,
    )

def build_reverse_native_keygen(target: Target) -> BuildResult:
    username = "analyst"
    serial = (sum((idx + 3) * ord(ch) for idx, ch in enumerate(username)) % 9000) + 1000
    flag = f"flag{{{serial}}}"
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        #include <stdlib.h>
        int main(void) {{
            int value = 0;
            if (scanf("%d", &value) != 1) return 1;
            if (value == {serial}) puts("ok");
            else puts("nope");
            return 0;
        }}
        """
    )
    binary, asm, _ = compile_reverse_bundle(source)
    notes = f"username={username}\nserial is 4 digits\n".encode("utf-8")
    statement = plain_statement(
        intro="附件是一份最小 keygenme，题面只给了用户名，要求恢复正确序列号。",
        steps=["结合样本逻辑恢复或验证序列号。", "把序列号按题面包装成 flag。", "提交结果。"],
        attachments=["challenge.bin", "challenge.asm", "notes.txt"],
    )
    solution = """# 解法

这题的输入空间只剩 4 位数字，黑盒验证也能做；正常做法是从汇编里恢复公式。本题正确序列号是 `%d`，所以提交 `flag{%d}`。
""" % (serial, serial)
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    for value in range(1000, 10000):
        proc = subprocess.run([binary], input=f"{value}\\n", text=True, capture_output=True)
        if "ok" in proc.stdout:
            print(f"flag{{{value}}}")
            return
    raise SystemExit("serial not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["题面给出的用户名会参与序列号计算。", "如果暂时没还原公式，4 位数字也能黑盒验证。"],
        files={
            "attachments/challenge.bin": binary,
            "attachments/challenge.asm": asm,
            "attachments/notes.txt": notes,
        },
        attachments=[
            ("attachments/challenge.bin", "challenge.bin"),
            ("attachments/challenge.asm", "challenge.asm"),
            ("attachments/notes.txt", "notes.txt"),
        ],
        flag_type="static",
        flag_value=flag,
    )

def build_reverse_native_string_table(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    key = 0x5C
    enc = [byte ^ key for byte in flag.encode("utf-8")]
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        static const unsigned char enc[] = {{{", ".join(str(v) for v in enc)}}};
        int main(void) {{
            puts("decode table first");
            return 0;
        }}
        """
    )
    binary, asm, rodata = compile_reverse_bundle(source)
    table = "\n".join(str(item) for item in enc).encode("utf-8")
    key_txt = f"{key}\n".encode("utf-8")
    statement = plain_statement(
        intro="附件里留了一份被简单异或的字符串表，真正的目标项就是 flag。",
        steps=["定位字符串表和 key。", "批量解密表项。", "恢复并提交 flag。"],
        attachments=["challenge.bin", "challenge.asm", "challenge.rodata.txt", "table.txt", "key.txt"],
    )
    solution = """# 解法

字符串表每个字节都只和同一个 key 做了 XOR。把表项逐个异或回去即可恢复目标字符串。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = [int(line) for line in (root / "attachments" / "table.txt").read_text(encoding="utf-8").splitlines() if line.strip()]
    key = int((root / "attachments" / "key.txt").read_text(encoding="utf-8").strip())
    print(bytes(value ^ key for value in values).decode())


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["这是最基础的一字节字符串表混淆。", "先把 key 找出来再批量逆运算。"],
        files={
            "attachments/challenge.bin": binary,
            "attachments/challenge.asm": asm,
            "attachments/challenge.rodata.txt": rodata,
            "attachments/table.txt": table,
            "attachments/key.txt": key_txt,
        },
        attachments=[
            ("attachments/challenge.bin", "challenge.bin"),
            ("attachments/challenge.asm", "challenge.asm"),
            ("attachments/challenge.rodata.txt", "challenge.rodata.txt"),
            ("attachments/table.txt", "table.txt"),
            ("attachments/key.txt", "key.txt"),
        ],
        flag_type="static",
        flag_value=flag,
    )

def build_reverse_js_mapper(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    arr = [hex(ord(ch) ^ 0x13) for ch in flag]
    script = textwrap.dedent(
        f"""\
        const data = [{", ".join(arr)}];
        function recover() {{
          return data.map(x => String.fromCharCode(parseInt(x, 16) ^ 0x13)).join('');
        }}
        console.log(recover() === process.argv[2] ? "ok" : "nope");
        """
    ).encode("utf-8")
    statement = plain_statement(
        intro="附件是一段轻量混淆的前端脚本，核心在数组映射和十六进制字符恢复。",
        steps=["拆掉数组映射。", "按逻辑恢复原字符串。", "提交得到的 flag。"],
        attachments=["check.js"],
    )
    solution = """# 解法

脚本并没有复杂控制流，关键就在 `map -> parseInt -> xor -> fromCharCode` 这条链。直接把数组里的值逆过去就能得到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "check.js").read_text(encoding="utf-8")
    raw = re.search(r"const data = \\[(.*?)\\];", text, re.S).group(1)
    values = [int(item.strip(), 16) for item in raw.split(",")]
    print("".join(chr(value ^ 0x13) for value in values))


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先把数组内容抠出来。", "真正的变换只有一次 xor。"],
        files={"attachments/check.js": script},
        attachments=[("attachments/check.js", "check.js")],
        flag_type="static",
        flag_value=flag,
    )

def build_reverse_shell_maze(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    shell = textwrap.dedent(
        f"""\
        #!/bin/sh
        pool="aa{flag}zz"
        left="${{pool#??}}"
        right="${{left%??}}"
        if [ "$1" = "$right" ]; then
          echo ok
        fi
        """
    ).encode("utf-8")
    statement = plain_statement(
        intro="附件是一段 shell crackme，重点是参数展开和前后缀裁剪。",
        steps=["理解 `${var#...}` 和 `${var%...}` 的行为。", "还原最终比较字符串。", "提交得到的 flag。"],
        attachments=["check.sh"],
    )
    solution = """# 解法

`${pool#??}` 去掉前两个字符，`${left%??}` 再去掉后两个字符。按这条链顺下来，最终比较值就是 flag 本身。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "check.sh").read_text(encoding="utf-8")
    pool = re.search(r'pool="([^"]+)"', text).group(1)
    left = pool[2:]
    right = left[:-2]
    print(right)


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["`#` 是去前缀，`%` 是去后缀。", "这里只用了最基础的两步裁剪。"],
        files={"attachments/check.sh": shell},
        attachments=[("attachments/check.sh", "check.sh")],
        flag_type="static",
        flag_value=flag,
    )

def build_reverse_vm_bytecode(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    program = bytearray()
    for ch in flag.encode("utf-8"):
        program.extend([0x01, ch ^ 0x21, 0x02, 0x21, 0x03])
    program.append(0xFF)
    vm = textwrap.dedent(
        """\
        # tiny vm
        # 0x01 imm -> push
        # 0x02 imm -> xor top
        # 0x03 -> output top
        # 0xFF -> halt
        """
    ).encode("utf-8")
    statement = plain_statement(
        intro="附件里给了一份简化虚拟机说明和一段字节码，需要自己解释执行才能恢复输出。",
        steps=["理解每条指令的语义。", "解释执行 program。", "把输出结果作为 flag 提交。"],
        attachments=["vm.txt", "program.vmb"],
    )
    solution = """# 解法

程序只有 push、xor、output、halt 四条指令。顺序解释执行即可，每组 `push -> xor -> output` 都会吐出一个明文字节。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    program = (root / "attachments" / "program.vmb").read_bytes()
    stack = []
    out = bytearray()
    pc = 0
    while pc < len(program):
        op = program[pc]
        pc += 1
        if op == 0x01:
            stack.append(program[pc])
            pc += 1
        elif op == 0x02:
            stack[-1] ^= program[pc]
            pc += 1
        elif op == 0x03:
            out.append(stack.pop())
        elif op == 0xFF:
            break
        else:
            raise SystemExit(f"unknown opcode {op}")
    print(out.decode())


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先把指令集读清楚。", "这类 VM 题通常先写个最小解释器。"],
        files={"attachments/vm.txt": vm, "attachments/program.vmb": bytes(program)},
        attachments=[("attachments/vm.txt", "vm.txt"), ("attachments/program.vmb", "program.vmb")],
        flag_type="static",
        flag_value=flag,
    )

def build_reverse_protocol_frame(target: Target) -> BuildResult:
    body = "10203040"
    flag = f"flag{{{body}}}"
    checker = textwrap.dedent(
        """\
        # frame: [a,b,c,d,sum,0x7e]
        expected = [0x10, 0x20, 0x30, 0x40]
        checksum = sum(expected) & 0xff
        """
    ).encode("utf-8")
    statement = plain_statement(
        intro="附件是一份简单协议校验脚本说明，需要构造合法帧并把关键字段按题面包装成 flag。",
        steps=["理解各字段含义。", "恢复前四个关键字节。", "按十六进制拼出 flag。"],
        attachments=["checker.txt"],
    )
    solution = """# 解法

校验条件已经把四个关键字节和校验方式给出来了。把前四个字节按两位十六进制拼起来，就是题目要求提交的 flag body。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import re


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "checker.txt").read_text(encoding="utf-8")
    raw = re.search(r"expected = \\[(.*?)\\]", text).group(1)
    nums = [int(item.strip(), 16) for item in raw.split(",")]
    print("flag{" + "".join(f"{value:02x}" for value in nums) + "}")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先区分数据字段和校验字段。", "题面只要求你恢复关键字段，不必真的发包。"],
        files={"attachments/checker.txt": checker},
        attachments=[("attachments/checker.txt", "checker.txt")],
        flag_type="static",
        flag_value=flag,
    )
