from __future__ import annotations

import re
from pathlib import Path

from docx import Document
from docx.shared import Pt


def count_non_whitespace_chars(text: str) -> int:
    return len(re.sub(r"\\s+", "", text))


def count_cjk_chars(text: str) -> int:
    # CJK Unified Ideographs + Extension A + Compatibility Ideographs
    pattern = r"[\u3400-\u4DBF\u4E00-\u9FFF\uF900-\uFAFF]"
    return len(re.findall(pattern, text))


def main() -> None:
    base_dir = Path(__file__).resolve().parent
    src_txt = base_dir / "Cyber Operations RangE (CORE)_ Containerized Gaming Platform for_中文学术直译.txt"
    out_docx = base_dir / "Cyber Operations RangE (CORE)_ Containerized Gaming Platform for_中文学术直译.docx"

    text = src_txt.read_text(encoding="utf-8")

    non_ws = count_non_whitespace_chars(text)
    cjk = count_cjk_chars(text)
    # 按常见“字数”口径优先检查中文字符数；同时保留非空白字符数用于调试。
    if cjk < 5000:
        raise SystemExit(
            f"翻译文本字数不足：cjk={cjk}, non_whitespace={non_ws}，未达到 5000 字要求"
        )

    doc = Document()

    style = doc.styles["Normal"]
    style.font.name = "SimSun"
    style.font.size = Pt(12)

    for line in text.splitlines():
        # 保持原文换行结构：每行写成一个段落，空行写空段落
        p = doc.add_paragraph(line.rstrip())
        p.paragraph_format.space_after = Pt(6)

    doc.save(out_docx)
    print(f"OK: wrote {out_docx}")
    print(f"counts: non_whitespace={non_ws}, cjk={cjk}")


if __name__ == "__main__":
    main()
