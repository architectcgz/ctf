from __future__ import annotations

from .helpers import *
from .models import BuildResult, Target

def build_forensics_browser_history(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    with tempfile.TemporaryDirectory() as tmp_dir:
        db_path = Path(tmp_dir) / "History.db"
        conn = sqlite3.connect(db_path)
        cur = conn.cursor()
        cur.executescript(
            """
            CREATE TABLE urls (
              id INTEGER PRIMARY KEY,
              url TEXT,
              title TEXT,
              visit_count INTEGER
            );
            """
        )
        rows = [
            ("https://academy.local/dashboard", "课程总览", 4),
            ("https://archive.local/tickets/night", "夜班工单", 2),
            (f"https://archive.local/redeem?flag={flag}", "sealed export", 1),
        ]
        cur.executemany("INSERT INTO urls(url, title, visit_count) VALUES(?,?,?)", rows)
        conn.commit()
        conn.close()
        db_bytes = db_path.read_bytes()
    statement = plain_statement(
        intro="附件是一份浏览器历史 sqlite 库，需要从访问记录里找出真正的导出地址。",
        steps=["查看表结构。", "筛选有价值的 URL 或标题。", "恢复并提交其中的 flag。"],
        attachments=["History.db"],
    )
    solution = """# 解法

浏览器历史本质上就是 sqlite。先看 `sqlite_master` 或直接查询 `urls` 表，再从 URL / title 中筛关键记录，能直接定位含 flag 的访问地址。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
import sqlite3
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    conn = sqlite3.connect(root / "attachments" / "History.db")
    cur = conn.cursor()
    rows = cur.execute("SELECT url, title FROM urls").fetchall()
    conn.close()
    text = "\\n".join(f"{url} {title}" for url, title in rows)
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
        hints=["先列 sqlite 表，不要一上来盲猜字段。", "浏览器历史最常看的就是 URL 和 title。"],
        files={"attachments/History.db": db_bytes},
        attachments=[("attachments/History.db", "History.db")],
        flag_type="static",
        flag_value=flag,
    )

def build_forensics_email_thread(target: Target) -> BuildResult:
    from email.message import EmailMessage

    flag = slug_flag(target.slug)
    msg = EmailMessage()
    msg["From"] = "audit@local"
    msg["To"] = "ops@local"
    msg["Subject"] = "Fwd: export summary"
    msg["X-Mailer"] = "Roundcube"
    msg.set_content("请看附件，真正的导出值不在正文里。")
    msg.add_attachment(
        f"sealed export code: {flag}\n".encode("utf-8"),
        maintype="text",
        subtype="plain",
        filename="note.txt",
    )
    eml_bytes = msg.as_bytes()
    statement = plain_statement(
        intro="一封转发邮件保留了 MIME 结构和附件编码，需要从样本里恢复真正的导出值。",
        steps=["解析邮件头和 MIME 边界。", "提取附件内容。", "提交附件中的 flag。"],
        attachments=["thread.eml"],
    )
    solution = """# 解法

用邮件解析器直接读 eml，遍历 multipart 的每个 part。正文只是提示，真正的值在附件 `note.txt` 里，解码后即可得到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

from email import policy
from email.parser import BytesParser
from pathlib import Path
import re


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    raw = (root / "attachments" / "thread.eml").read_bytes()
    msg = BytesParser(policy=policy.default).parsebytes(raw)
    text_parts = []
    for part in msg.walk():
        if part.get_content_maintype() == "multipart":
            continue
        payload = part.get_payload(decode=True) or b""
        text_parts.append(payload.decode("utf-8", errors="ignore"))
    text = "\\n".join(text_parts)
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
        hints=["正文和附件是不同 part。", "先把 MIME 结构走通，再看具体内容。"],
        files={"attachments/thread.eml": eml_bytes},
        attachments=[("attachments/thread.eml", "thread.eml")],
        flag_type="static",
        flag_value=flag,
    )

def build_forensics_authlog(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    log_text = textwrap.dedent(
        f"""\
        May 09 09:10:10 lab sshd[1001]: Failed password for invalid user test from 10.0.0.3 port 50122 ssh2
        May 09 09:10:11 lab sshd[1001]: Failed password for invalid user test from 10.0.0.3 port 50122 ssh2
        May 09 09:12:44 lab sshd[1099]: Accepted password for deploy from 10.0.0.3 port 50148 ssh2
        May 09 09:12:55 lab sudo: deploy : TTY=pts/0 ; PWD=/srv ; USER=root ; COMMAND=/bin/cat /srv/{flag}.txt
        May 09 09:13:08 lab CRON[1120]: (root) CMD (/usr/local/bin/rotate-backup)
        """
    )
    statement = plain_statement(
        intro="附件是一段认证日志，需要还原成功登录后的关键操作。",
        steps=["按时间线区分失败登录和真正成功事件。", "找到成功后的敏感命令。", "从命令痕迹中恢复 flag。"],
        attachments=["auth.log"],
    )
    solution = """# 解法

先按时间顺序区分爆破噪声和成功登录，再看成功登录后的 `sudo` / `COMMAND=` 记录。本题的敏感操作直接读了带 flag 名称的文件。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "auth.log").read_text(encoding="utf-8")
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
        hints=["先看 Accepted password 之后发生了什么。", "sudo COMMAND 行通常最有价值。"],
        files={"attachments/auth.log": log_text.encode("utf-8")},
        attachments=[("attachments/auth.log", "auth.log")],
        flag_type="static",
        flag_value=flag,
    )

def build_forensics_pcap_basic_auth(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    credential = base64.b64encode(f"ops:{flag}".encode("utf-8"))
    http = (
        b"GET /private/export HTTP/1.1\r\n"
        b"Host: intranet.local\r\n"
        b"User-Agent: curl/8.5\r\n"
        b"Authorization: Basic " + credential + b"\r\n\r\n"
    )
    src_ip = b"\x0a\x00\x00\x03"
    dst_ip = b"\x0a\x00\x00\x09"
    tcp = build_tcp_segment(src_ip, dst_ip, 50001, 80, http)
    pcap = build_pcap([wrap_ethernet(build_ipv4_packet(src_ip, dst_ip, 6, tcp))])
    statement = plain_statement(
        intro="附件是一段 HTTP 抓包，需要从认证头里恢复真正的访问凭据。",
        steps=["定位含认证信息的请求。", "解码 Basic Auth 头。", "恢复并提交其中的 flag。"],
        attachments=["capture.pcap"],
    )
    solution = """# 解法

直接在抓包里找 `Authorization: Basic ...`。把 Base64 解开后会得到 `username:password`，本题把 flag 放在 password 位置。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import base64
import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = (root / "attachments" / "capture.pcap").read_bytes()
    match = re.search(rb"Authorization: Basic ([A-Za-z0-9+/=]+)", data)
    if not match:
        raise SystemExit("auth header not found")
    plain = base64.b64decode(match.group(1)).decode()
    print(plain.split(":", 1)[1])


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["HTTP 明文协议里的认证头通常直接可见。", "Basic Auth 只是 Base64，不是加密。"],
        files={"attachments/capture.pcap": pcap},
        attachments=[("attachments/capture.pcap", "capture.pcap")],
        flag_type="static",
        flag_value=flag,
    )

def build_forensics_pcap_dns_exfil(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    encoded = base64.b32encode(flag.encode("utf-8")).decode("ascii").rstrip("=")
    chunks = [encoded[idx : idx + 10] for idx in range(0, len(encoded), 10)]
    frames = []
    src_ip = b"\x0a\x00\x00\x04"
    dst_ip = b"\x0a\x00\x00\x35"
    for idx, chunk in enumerate(chunks):
        qname = f"{idx:02d}.{chunk}.exfil.lab"
        dns = build_dns_query(qname)
        udp = build_udp_datagram(src_ip, dst_ip, 53000 + idx, 53, dns)
        frames.append(wrap_ethernet(build_ipv4_packet(src_ip, dst_ip, 17, udp)))
    pcap = build_pcap(frames)
    statement = plain_statement(
        intro="附件是一段 DNS 抓包，怀疑有人把数据拆成子域名做外带。",
        steps=["定位可疑查询名。", "按序拼接子域名里的数据分片。", "解码后提交 flag。"],
        attachments=["dns.pcap"],
    )
    solution = """# 解法

抓包里每个 DNS Query Name 都带了一个编号和一段 Base32 数据。按编号排序后拼起来，再补齐 `=` 做 Base32 解码即可恢复 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import base64
import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = (root / "attachments" / "dns.pcap").read_bytes()
    parts = re.findall(rb"\\x02(\\d\\d)[\\x01-\\x0a]([A-Z2-7]{1,10})\\x05exfil\\x03lab\\x00", data)
    ordered = [chunk.decode() for _, chunk in sorted(parts, key=lambda item: item[0])]
    joined = "".join(ordered)
    padded = joined + "=" * ((8 - len(joined) % 8) % 8)
    print(base64.b32decode(padded).decode())


if __name__ == "__main__":
    main()
"""
    return BuildResult(
        statement=statement,
        solution=solution,
        solve_py=solve_py,
        hints=["子域名前面的两位编号不要丢。", "大写字母加 2-7 很像 Base32。"],
        files={"attachments/dns.pcap": pcap},
        attachments=[("attachments/dns.pcap", "dns.pcap")],
        flag_type="static",
        flag_value=flag,
    )

def build_forensics_git_reflog(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    with tempfile.TemporaryDirectory() as tmp_dir:
        repo = Path(tmp_dir) / "repo"
        repo.mkdir()
        subprocess.run(["git", "init"], cwd=repo, check=True, stdout=subprocess.DEVNULL)
        subprocess.run(["git", "config", "user.name", "CTF"], cwd=repo, check=True)
        subprocess.run(["git", "config", "user.email", "ctf@example.local"], cwd=repo, check=True)
        (repo / "README.md").write_text("repo snapshot\n", encoding="utf-8")
        subprocess.run(["git", "add", "README.md"], cwd=repo, check=True)
        subprocess.run(["git", "commit", "-m", "init"], cwd=repo, check=True, stdout=subprocess.DEVNULL)
        (repo / "draft.txt").write_text(f"stash this secret: {flag}\n", encoding="utf-8")
        subprocess.run(["git", "add", "draft.txt"], cwd=repo, check=True)
        subprocess.run(["git", "stash", "push", "-m", "night-draft"], cwd=repo, check=True, stdout=subprocess.DEVNULL)
        archive = build_tar_gz(
            {
                str(path.relative_to(repo)): path.read_bytes()
                for path in repo.rglob("*")
                if path.is_file()
            }
        )
    statement = plain_statement(
        intro="附件是一份带 `.git` 历史的仓库快照，需要从 stash / reflog 残留里恢复草稿内容。",
        steps=["解压仓库样本。", "检查 stash、reflog 或悬挂对象。", "恢复草稿中的 flag。"],
        attachments=["repo.tar.gz"],
    )
    solution = """# 解法

样本不是普通源码包，而是完整仓库快照。进入目录后直接看 `git stash list` 和 `git stash show -p`，即可从暂存草稿里恢复 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
import shutil
import subprocess
import tarfile
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        tar_path = root / "attachments" / "repo.tar.gz"
        with tarfile.open(tar_path, "r:gz") as tf:
            tf.extractall(tmp_dir)
        repo = Path(tmp_dir)
        output = subprocess.check_output(["git", "stash", "show", "-p"], cwd=repo, text=True)
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
        hints=["不要只看工作区文件，重点在 `.git`。", "stash 本身就是一条很直接的线索。"],
        files={"attachments/repo.tar.gz": archive},
        attachments=[("attachments/repo.tar.gz", "repo.tar.gz")],
        flag_type="static",
        flag_value=flag,
    )

def build_forensics_office_review(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    docx = build_zip(
        {
            "[Content_Types].xml": b"""<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/comments.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.comments+xml"/></Types>""",
            "word/document.xml": b"""<?xml version="1.0" encoding="UTF-8"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>review draft</w:t></w:r></w:p></w:body></w:document>""",
            "word/comments.xml": f"""<?xml version="1.0" encoding="UTF-8"?><w:comments xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:comment w:id="0"><w:p><w:r><w:t>{flag}</w:t></w:r></w:p></w:comment></w:comments>""".encode(
                "utf-8"
            ),
        }
    )
    statement = plain_statement(
        intro="附件是一份 OOXML 文档，表面内容很少，但修订和批注痕迹没有清干净。",
        steps=["把文档当 zip 包处理。", "检查 `word/` 下的 XML。", "从批注或修订痕迹中恢复 flag。"],
        attachments=["review.docx"],
    )
    solution = """# 解法

`.docx` 本质就是 zip。解开后重点看 `word/comments.xml`、`word/document.xml`、`word/footnotes.xml` 等结构化 XML，本题的 flag 直接留在 comments 里。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path
from zipfile import ZipFile


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with ZipFile(root / "attachments" / "review.docx") as zf:
        text = zf.read("word/comments.xml").decode("utf-8")
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
        hints=["OOXML 文档本质是 zip。", "优先检查 comments、revisions 这类侧边结构。"],
        files={"attachments/review.docx": docx},
        attachments=[("attachments/review.docx", "review.docx")],
        flag_type="static",
        flag_value=flag,
    )

def build_forensics_sqlite_wal(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    with tempfile.TemporaryDirectory() as tmp_dir:
        db_path = Path(tmp_dir) / "chat.db"
        conn = sqlite3.connect(db_path)
        cur = conn.cursor()
        cur.execute("PRAGMA journal_mode=WAL;")
        cur.execute("PRAGMA wal_autocheckpoint=0;")
        cur.execute("CREATE TABLE messages(id INTEGER PRIMARY KEY, sender TEXT, body TEXT);")
        cur.execute("INSERT INTO messages(sender, body) VALUES (?, ?)", ("ops", "normal note"))
        cur.execute("INSERT INTO messages(sender, body) VALUES (?, ?)", ("lead", f"sealed line {flag}"))
        conn.commit()
        wal_path = Path(str(db_path) + "-wal")
        shm_path = Path(str(db_path) + "-shm")
        db_bytes = db_path.read_bytes()
        wal_bytes = wal_path.read_bytes()
        shm_bytes = shm_path.read_bytes()
        conn.close()
    statement = plain_statement(
        intro="聊天样本里除了主库，还附带了 WAL 和 SHM 文件，需要按完整现场恢复记录。",
        steps=["不要只看主库，连同 `-wal` 和 `-shm` 一起分析。", "查询消息记录。", "恢复并提交其中的 flag。"],
        attachments=["chat.db", "chat.db-wal", "chat.db-shm"],
    )
    solution = """# 解法

sqlite 在 WAL 模式下，最近的提交可能只落在 `-wal` 里。把主库和 sidecar 文件放在同目录后直接用 sqlite 打开查询，就能看到 WAL 中的记录。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
import shutil
import sqlite3
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        tmp = Path(tmp_dir)
        for name in ("chat.db", "chat.db-wal", "chat.db-shm"):
            shutil.copy2(root / "attachments" / name, tmp / name)
        conn = sqlite3.connect(tmp / "chat.db")
        rows = conn.execute("SELECT sender, body FROM messages").fetchall()
        conn.close()
    text = "\\n".join(f"{sender}: {body}" for sender, body in rows)
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
        hints=["把三份文件放在同一目录再开。", "这是典型的 WAL 残留场景。"],
        files={
            "attachments/chat.db": db_bytes,
            "attachments/chat.db-wal": wal_bytes,
            "attachments/chat.db-shm": shm_bytes,
        },
        attachments=[
            ("attachments/chat.db", "chat.db"),
            ("attachments/chat.db-wal", "chat.db-wal"),
            ("attachments/chat.db-shm", "chat.db-shm"),
        ],
        flag_type="static",
        flag_value=flag,
    )

def build_forensics_registry(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    reg_text = textwrap.dedent(
        f"""\
        Windows Registry Editor Version 5.00

        [HKEY_CURRENT_USER\\Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\RunMRU]
        "a"="notepad C:\\\\Users\\\\ops\\\\Desktop\\\\todo.txt"
        "b"="cmd /c type C:\\\\vault\\\\{flag}.txt"
        "MRUList"="ba"
        """
    )
    statement = plain_statement(
        intro="附件是一份注册表导出，需要从 RunMRU 执行记录里恢复最关键的一次命令。",
        steps=["读取导出文本。", "定位 RunMRU 项。", "从执行命令痕迹中提取 flag。"],
        attachments=["runmru.reg"],
    )
    solution = """# 解法

RunMRU 直接记录了用户在运行框执行过的命令，`MRUList` 还能标出先后顺序。本题里命令参数本身就暴露了 flag 文件名。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "runmru.reg").read_text(encoding="utf-8")
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
        hints=["RunMRU 是最直接的命令执行痕迹之一。", "先定位这个键，再看值内容。"],
        files={"attachments/runmru.reg": reg_text.encode("utf-8")},
        attachments=[("attachments/runmru.reg", "runmru.reg")],
        flag_type="static",
        flag_value=flag,
    )

def build_forensics_memory_snapshot(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    rng = rng_for(target.slug)
    noise = bytes(rng.getrandbits(8) for _ in range(256))
    dump = noise + b"USER=ops\x00MODE=debug\x00" + f"RECOVERY_CODE={flag}\x00".encode("utf-8") + noise[::-1]
    statement = plain_statement(
        intro="附件是一段原始内存快照，没有结构化元数据，只能先从字符串线索入手。",
        steps=["对原始样本做字符串提取。", "筛出类似环境变量的内容。", "恢复并提交 flag。"],
        attachments=["snapshot.bin"],
    )
    solution = """# 解法

这类无结构快照先跑 `strings` 是最稳的入口。提取结果里有多条环境变量样式的键值对，直接能定位 `RECOVERY_CODE=...`。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = (root / "attachments" / "snapshot.bin").read_bytes()
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
        hints=["先用 strings 类思路筛文本。", "环境变量样式的键值对很值得重点看。"],
        files={"attachments/snapshot.bin": dump},
        attachments=[("attachments/snapshot.bin", "snapshot.bin")],
        flag_type="static",
        flag_value=flag,
    )

def build_forensics_docker_layer(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    layer1 = build_tar_gz({"app/.env": f"EXPORT_FLAG={flag}\n".encode("utf-8")})
    layer2 = build_tar_gz({"app/.wh..env": b""})
    image_tar = build_zip(
        {
            "manifest.json": json.dumps([{"Config": "config.json", "RepoTags": ["demo:latest"], "Layers": ["layer1.tar.gz", "layer2.tar.gz"]}]).encode("utf-8"),
            "config.json": json.dumps({"history": [{"created_by": "COPY .env /app/.env"}, {"created_by": "RUN rm /app/.env"}]}).encode("utf-8"),
            "layer1.tar.gz": layer1,
            "layer2.tar.gz": layer2,
        }
    )
    statement = plain_statement(
        intro="附件是一份镜像导出包，需要顺着 layer 历史去找已经被删除的敏感文件。",
        steps=["查看 manifest 和 config。", "逐层展开 layer。", "从早期层残留中恢复 flag。"],
        attachments=["image.zip"],
    )
    solution = """# 解法

镜像后续层虽然删除了 `.env`，但早期 layer 里仍保留完整文件内容。看 `manifest.json` 的 layer 顺序，再展开第一层就能直接看到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import io
import re
import tarfile
from pathlib import Path
from zipfile import ZipFile


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with ZipFile(root / "attachments" / "image.zip") as zf:
        layer = zf.read("layer1.tar.gz")
    with tarfile.open(fileobj=io.BytesIO(layer), mode="r:gz") as tf:
        data = tf.extractfile("app/.env").read().decode("utf-8")
    match = re.search(r"flag\\{[a-z0-9_\\-]+\\}", data)
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
        hints=["后续删除不会抹掉早期层内容。", "先看 manifest 再决定读哪一层。"],
        files={"attachments/image.zip": image_tar},
        attachments=[("attachments/image.zip", "image.zip")],
        flag_type="static",
        flag_value=flag,
    )
