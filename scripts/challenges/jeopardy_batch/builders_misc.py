from __future__ import annotations

from .helpers import *
from .models import BuildResult, Target

def build_misc_special_filename(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    archive = build_tar_gz({"./-": f"{flag}\n".encode("utf-8")})
    statement = plain_statement(
        intro="附件是一份目录样本，真正的目标文件名会被 shell 当成参数解释。",
        steps=["解压样本。", "处理特殊文件名。", "读取内容并提交 flag。"],
        attachments=["sample.tar.gz"],
    )
    solution = """# 解法

文件名本身就是 `-`，直接 `cat -` 会被 shell 当成标准输入。加上 `./` 或 `--` 之后再读，就能拿到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import tarfile
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        with tarfile.open(root / "attachments" / "sample.tar.gz", "r:gz") as tf:
            tf.extractall(tmp_dir)
        print((Path(tmp_dir) / "-").read_text(encoding="utf-8").strip())


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先想想 shell 怎么区分参数和路径。", "前缀 `./` 通常能解决这类问题。"],
        files={"attachments/sample.tar.gz": archive},
        attachments=[("attachments/sample.tar.gz", "sample.tar.gz")],
        flag_type="static",
        flag_value=flag,
    )

def build_misc_find_target_file(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    entries = {
        "tree/a.txt": b"hello\n",
        "tree/bin/run.sh": b"#!/bin/sh\necho run\n",
        "tree/logs/latest.log": b"ok\n",
        "tree/vault/target.bin": flag.encode("utf-8"),
    }
    archive = build_tar_gz(entries, mode_map={"tree/bin/run.sh": 0o755, "tree/vault/target.bin": 0o755})
    statement = plain_statement(
        intro="附件里有一大堆文件，只有一个目标文件同时满足“可执行、位于 vault、大小刚好等于 flag 长度”。",
        steps=["解压目录树。", "按路径、权限和大小筛文件。", "读取目标内容并提交 flag。"],
        attachments=["tree.tar.gz"],
    )
    solution = """# 解法

用 `find` 配合 `-path`、`-perm`、`-size` 就能很快收口到目标文件。本题唯一满足条件的是 `vault/target.bin`。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import tarfile
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        with tarfile.open(root / "attachments" / "tree.tar.gz", "r:gz") as tf:
            tf.extractall(tmp_dir)
        for path in Path(tmp_dir).rglob("*"):
            if path.is_file() and "vault" in path.parts and (path.stat().st_mode & 0o111):
                text = path.read_text(encoding="utf-8")
                if text.startswith("flag{"):
                    print(text)
                    return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["条件组合查找比一个个点开快得多。", "重点是路径、权限和大小同时收口。"],
        files={"attachments/tree.tar.gz": archive},
        attachments=[("attachments/tree.tar.gz", "tree.tar.gz")],
        flag_type="static",
        flag_value=flag,
    )

def build_misc_strings_signal(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    rng = rng_for(target.slug)
    blob = bytearray(rng.getrandbits(8) for _ in range(512))
    blob.extend(b"random status line\x00")
    blob.extend(f"sealed_marker={flag}\x00".encode("utf-8"))
    blob.extend(bytes(rng.getrandbits(8) for _ in range(256)))
    statement = plain_statement(
        intro="附件是一份混杂二进制转储，真正线索只在可打印字符串里短暂出现。",
        steps=["先提取可打印字符串。", "再按关键字筛选。", "找到并提交 flag。"],
        attachments=["blob.bin"],
    )
    solution = """# 解法

这种题先跑 `strings`。提取后的文本里有很多噪声，但 `marker` / `sealed` 这类关键词很容易把包含 flag 的那一行筛出来。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = (root / "attachments" / "blob.bin").read_bytes()
    match = re.search(rb"flag\\{[a-z0-9_\\-]+\\}", data)
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
        hints=["不要先猜格式，先从 strings 入手。", "关键词筛选通常能快速去掉大量噪声。"],
        files={"attachments/blob.bin": bytes(blob)},
        attachments=[("attachments/blob.bin", "blob.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_misc_hexdump_chain(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    final_tar = build_tar_gz({"flag.txt": f"{flag}\n".encode("utf-8")})
    packed = bz2.compress(gzip.compress(final_tar))
    hexdump = "\n".join(packed.hex()[idx : idx + 64] for idx in range(0, len(packed.hex()), 64))
    statement = plain_statement(
        intro="附件不是直接的二进制，而是一份十六进制转储；还原后还有多层压缩归档。",
        steps=["把十六进制转回原始字节流。", "按实际文件头逐层解压。", "取出最终 flag。"],
        attachments=["dump.txt"],
    )
    solution = """# 解法

先把 hexdump 还原为原始字节，再根据 `file` / 魔数判断是 `bzip2 -> gzip -> tar.gz` 这条链。逐层拆开后就能读到 `flag.txt`。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import bz2
import gzip
import io
import tarfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "dump.txt").read_text(encoding="utf-8")
    data = bytes.fromhex("".join(text.split()))
    stage1 = bz2.decompress(data)
    stage2 = gzip.decompress(stage1)
    with tarfile.open(fileobj=io.BytesIO(stage2), mode="r:gz") as tf:
        print(tf.extractfile("flag.txt").read().decode().strip())


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先别急着脚本化，先看文件头。", "每一层都能从魔数判断出类型。"],
        files={"attachments/dump.txt": hexdump.encode("utf-8")},
        attachments=[("attachments/dump.txt", "dump.txt")],
        flag_type="static",
        flag_value=flag,
    )

def build_misc_png_text(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    raw = b"\x89PNG\r\n\x1a\n"
    ihdr = struct.pack("!IIBBBBB", 1, 1, 8, 2, 0, 0, 0)
    idat = zlib.compress(b"\x00\x00\x00\x00")
    text_data = f"note={base64.b64encode(flag.encode()).decode()}".encode("utf-8")
    png = raw + png_chunk(b"IHDR", ihdr) + png_chunk(b"tEXt", text_data) + png_chunk(b"IDAT", idat) + png_chunk(b"IEND", b"")
    statement = plain_statement(
        intro="附件是一张看起来很普通的 PNG 图片，但线索不在像素本身。",
        steps=["检查 PNG chunk。", "提取文本块内容。", "解码并提交 flag。"],
        attachments=["note.png"],
    )
    solution = """# 解法

PNG 的 `tEXt` / `zTXt` / `iTXt` 都可能藏信息。本题的 `tEXt` 块里是一段 Base64，解开后就是 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import base64
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = (root / "attachments" / "note.png").read_bytes()
    cursor = 8
    while cursor < len(data):
        length = int.from_bytes(data[cursor:cursor+4], "big")
        ctype = data[cursor+4:cursor+8]
        body = data[cursor+8:cursor+8+length]
        cursor += 12 + length
        if ctype == b"tEXt":
            _, value = body.split(b"=", 1)
            print(base64.b64decode(value).decode())
            return
    raise SystemExit("tEXt not found")


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["PNG 不是只有像素数据。", "文本 chunk 常见于元数据或调试信息。"],
        files={"attachments/note.png": png},
        attachments=[("attachments/note.png", "note.png")],
        flag_type="static",
        flag_value=flag,
    )

def build_misc_pdf_embedded(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    pdf = textwrap.dedent(
        f"""\
        %PDF-1.4
        1 0 obj
        << /Type /Catalog /Names << /EmbeddedFiles << /Names [(note.txt) 2 0 R] >> >> >>
        endobj
        2 0 obj
        << /Type /Filespec /F (note.txt) /EF << /F 3 0 R >> >>
        endobj
        3 0 obj
        << /Type /EmbeddedFile /Length {len(flag) + 1} >>
        stream
        {flag}
        endstream
        endobj
        trailer << /Root 1 0 R >>
        %%EOF
        """
    ).encode("utf-8")
    statement = plain_statement(
        intro="附件是一份 PDF，题面正文几乎没信息，线索藏在内嵌对象里。",
        steps=["检查 PDF 对象结构。", "定位 EmbeddedFile / stream。", "恢复并提交 flag。"],
        attachments=["brief.pdf"],
    )
    solution = """# 解法

PDF 可以内嵌文件或直接把数据放进 stream。查对象结构后定位 `EmbeddedFile` 对象，读出对应 stream 就能拿到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "brief.pdf").read_text(encoding="utf-8")
    match = re.search(r"flag\\{[a-z0-9_\\-]+\\}", text)
    if not match:
        raise SystemExit("flag not found")
    print(match.group(0))


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["重点看 `/EmbeddedFile` 和 `stream`。", "先把 PDF 当结构化文本对象看。"],
        files={"attachments/brief.pdf": pdf},
        attachments=[("attachments/brief.pdf", "brief.pdf")],
        flag_type="static",
        flag_value=flag,
    )

def build_misc_bmp_lsb(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    bits = "".join(f"{byte:08b}" for byte in (flag + "\x00").encode("utf-8"))
    width = len(bits)
    row_size = ((width * 3 + 3) // 4) * 4
    pixel_data = bytearray()
    for bit in bits:
        blue = 0xAA | int(bit)
        pixel_data.extend([blue, 0x55, 0x11])
    pixel_data.extend(b"\x00" * (row_size - width * 3))
    file_size = 54 + len(pixel_data)
    header = bytearray()
    header.extend(b"BM")
    header.extend(struct.pack("<I", file_size))
    header.extend(b"\x00\x00\x00\x00")
    header.extend(struct.pack("<I", 54))
    header.extend(struct.pack("<IiiHHIIiiII", 40, width, 1, 1, 24, 0, len(pixel_data), 2835, 2835, 0, 0))
    bmp = bytes(header) + bytes(pixel_data)
    statement = plain_statement(
        intro="附件是一张 BMP 图片，表面像素没什么可读内容，但最低位被动过手脚。",
        steps=["解析 BMP 像素数组。", "提取固定通道的 LSB。", "重组文本并提交 flag。"],
        attachments=["cover.bmp"],
    )
    solution = """# 解法

BMP 像素数据没有压缩，直接读蓝色通道最低位即可。按 8 位一组恢复字节，直到读到终止符，就能拿到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = (root / "attachments" / "cover.bmp").read_bytes()
    offset = int.from_bytes(data[10:14], "little")
    width = int.from_bytes(data[18:22], "little", signed=True)
    row_size = ((width * 3 + 3) // 4) * 4
    row = data[offset : offset + row_size]
    bits = "".join(str(row[idx * 3] & 1) for idx in range(width))
    out = bytearray()
    for idx in range(0, len(bits), 8):
        byte = int(bits[idx : idx + 8], 2)
        if byte == 0:
            break
        out.append(byte)
    print(out.decode())


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["BMP 最适合做直接像素位分析。", "优先看未压缩的单个颜色通道。"],
        files={"attachments/cover.bmp": bmp},
        attachments=[("attachments/cover.bmp", "cover.bmp")],
        flag_type="static",
        flag_value=flag,
    )

def build_misc_cron_spool(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    archive = build_tar_gz(
        {
            "etc/crontab": b"SHELL=/bin/sh\n* * * * * root run-parts /etc/cron.hourly\n",
            "var/spool/cron/root": b"*/15 * * * * /usr/local/bin/night-export.sh\n",
            "usr/local/bin/night-export.sh": f"#!/bin/sh\nOUTPUT=/srv/{flag}.txt\nprintf 'sealed' > \"$OUTPUT\"\n".encode("utf-8"),
        },
        mode_map={"usr/local/bin/night-export.sh": 0o755},
    )
    statement = plain_statement(
        intro="附件保留了一段计划任务现场，需要顺着 cron 链找到真正的夜间导出目标。",
        steps=["从 crontab 找到执行入口。", "跟进到脚本内容。", "恢复脚本里写出的 flag。"],
        attachments=["cron.tar.gz"],
    )
    solution = """# 解法

先看 `var/spool/cron/root`，能直接定位实际执行的脚本。再读脚本内容，就能在 `OUTPUT=` 变量里看到带 flag 名称的目标路径。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
import tarfile
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        with tarfile.open(root / "attachments" / "cron.tar.gz", "r:gz") as tf:
            tf.extractall(tmp_dir)
        script = (Path(tmp_dir) / "usr/local/bin/night-export.sh").read_text(encoding="utf-8")
    match = re.search(r"flag\\{[a-z0-9_\\-]+\\}", script)
    if not match:
        raise SystemExit("flag not found")
    print(match.group(0))


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["先找入口，再跟脚本链，不要只停在 crontab。", "输出路径变量通常就是关键线索。"],
        files={"attachments/cron.tar.gz": archive},
        attachments=[("attachments/cron.tar.gz", "cron.tar.gz")],
        flag_type="static",
        flag_value=flag,
    )

def build_misc_makefile_override(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    body = flag.removeprefix("flag{").removesuffix("}")
    part1, part2 = body[: len(body) // 2], body[len(body) // 2 :]
    package = build_tar_gz(
        {
            "Makefile": textwrap.dedent(
                """\
                include rules.mk
                ROLE ?= guest
                reveal:
                \t@if [ "$(ROLE)" = "admin" ]; then \\
                \t\tprintf '%s%s%s\\n' "$(PART1)" "$(PART2)" "$(PART3)"; \\
                \telse \\
                \t\techo denied; \\
                \tfi
                """
            ).encode("utf-8"),
            "rules.mk": f"PART1=flag{{{part1}\nPART2={part2}\nPART3=}}\n".encode("utf-8"),
        }
    )
    statement = plain_statement(
        intro="附件是一份最小构建工程，真正的输出需要理解 Make 变量覆盖和条件分支。",
        steps=["阅读 Makefile 和 include 文件。", "用正确变量触发 reveal 目标。", "拿到输出中的 flag。"],
        attachments=["makepack.tar.gz"],
    )
    solution = """# 解法

`ROLE ?= guest` 说明命令行变量可以覆盖默认值。运行 `make reveal ROLE=admin` 就会进入输出分支，拼出真正的 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import tarfile
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        with tarfile.open(root / "attachments" / "makepack.tar.gz", "r:gz") as tf:
            tf.extractall(tmp_dir)
        output = subprocess.check_output(["make", "reveal", "ROLE=admin"], cwd=tmp_dir, text=True)
    print(output.strip())


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["注意 `?=` 的含义。", "命令行变量优先级高于默认值。"],
        files={"attachments/makepack.tar.gz": package},
        attachments=[("attachments/makepack.tar.gz", "makepack.tar.gz")],
        flag_type="static",
        flag_value=flag,
    )

def build_misc_terminal_cast(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    cast = "\n".join(
        [
            '{"version": 2, "width": 80, "height": 24, "timestamp": 1715251200, "env": {"TERM": "xterm-256color"}}',
            '[0.1, "o", "$ cat /srv/export-code.txt\\r\\n"]',
            f'[0.2, "o", "{flag}\\r\\n"]',
        ]
    ).encode("utf-8")
    statement = plain_statement(
        intro="附件是一段终端录屏日志，需要从回放事件里还原实际输出。",
        steps=["识别录屏日志格式。", "读取输出事件。", "恢复并提交其中的 flag。"],
        attachments=["session.cast"],
    )
    solution = """# 解法

Asciinema v2 的第一行是头信息，后续每行都是 JSON 事件。直接筛 `\"o\"` 输出事件就能看到终端真正打印的内容，其中包含 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    lines = (root / "attachments" / "session.cast").read_text(encoding="utf-8").splitlines()[1:]
    output = "".join(json.loads(line)[2] for line in lines)
    match = re.search(r"flag\\{[a-z0-9_\\-]+\\}", output)
    if not match:
        raise SystemExit("flag not found")
    print(match.group(0))


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["第一行是头，不是事件。", "输出事件一般用 `o` 标识。"],
        files={"attachments/session.cast": cast},
        attachments=[("attachments/session.cast", "session.cast")],
        flag_type="static",
        flag_value=flag,
    )

def build_misc_env_merge(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    archive = build_tar_gz(
        {
            ".env": b"EXPORT_CODE=base\n",
            "service.env": b"EXPORT_CODE=service\n",
            "compose.env": f"EXPORT_CODE={flag}\n".encode("utf-8"),
            "README.txt": b"priority: compose.env > service.env > .env\n",
        }
    )
    statement = plain_statement(
        intro="附件里有多份环境配置，目标是按题面给的优先级推导最终生效值。",
        steps=["读取各层配置文件。", "按优先级覆盖同名变量。", "恢复最终的 EXPORT_CODE 并按 flag 提交。"],
        attachments=["envpack.tar.gz"],
    )
    solution = """# 解法

这类题不需要盲猜，只要按优先级顺序做最后一次覆盖即可。本题最终生效的是 `compose.env` 里的 `EXPORT_CODE`，它本身就是 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import tarfile
import tempfile
from pathlib import Path


def load_env(path: Path) -> dict[str, str]:
    values = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        if "=" in line:
            key, value = line.split("=", 1)
            values[key] = value
    return values


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        with tarfile.open(root / "attachments" / "envpack.tar.gz", "r:gz") as tf:
            tf.extractall(tmp_dir)
        tmp = Path(tmp_dir)
        merged = {}
        for name in (".env", "service.env", "compose.env"):
            merged.update(load_env(tmp / name))
    print(merged["EXPORT_CODE"])


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["这题关键不在格式，而在覆盖顺序。", "最终值就是最后一个生效的同名变量。"],
        files={"attachments/envpack.tar.gz": archive},
        attachments=[("attachments/envpack.tar.gz", "envpack.tar.gz")],
        flag_type="static",
        flag_value=flag,
    )
