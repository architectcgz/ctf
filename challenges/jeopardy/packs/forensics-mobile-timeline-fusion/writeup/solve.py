#!/usr/bin/env python3
from __future__ import annotations

import base64
import csv
import json
import re
import sqlite3
import tempfile
from datetime import datetime, timezone
from pathlib import Path
from zipfile import ZipFile


INCIDENT_RE = re.compile(
    r"incident=(?P<tag>[a-z0-9_-]+)\s+window_start=(?P<start>\S+)\s+window_end=(?P<end>\S+)\s+assemble=utc"
)


def parse_ts(value: str) -> datetime:
    text = value.strip()
    if text.endswith("Z"):
        text = text[:-1] + "+00:00"
    return datetime.fromisoformat(text).astimezone(timezone.utc)


def decode_fragment(kind: str, payload: str) -> str:
    if kind == "base64":
        return base64.b64decode(payload).decode("utf-8")
    if kind == "rot13":
        return payload.translate(str.maketrans(
            "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
            "NOPQRSTUVWXYZABCDEFGHIJKLMnopqrstuvwxyzabcdefghijklm",
        ))
    if kind == "hex":
        return bytes.fromhex(payload).decode("utf-8")
    if kind == "reverse":
        return payload[::-1]
    raise SystemExit(f"unknown encoding: {kind}")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    archive = root / "attachments" / "case-evidence.zip"

    events: list[tuple[datetime, str]] = []
    with ZipFile(archive) as zf, tempfile.TemporaryDirectory() as tmp_dir:
        chat_path = Path(tmp_dir) / "chat.db"
        chat_path.write_bytes(zf.read("mobile/chat.db"))

        conn = sqlite3.connect(chat_path)
        row = conn.execute(
            "SELECT body FROM messages WHERE deleted = 1 AND body LIKE 'incident=%' LIMIT 1"
        ).fetchone()
        conn.close()
        if row is None:
            raise SystemExit("incident note not found")

        match = INCIDENT_RE.search(row[0])
        if not match:
            raise SystemExit("incident note format invalid")

        tag = match.group("tag")
        start = parse_ts(match.group("start"))
        end = parse_ts(match.group("end"))

        clipboard = json.loads(zf.read("mobile/clipboard.json").decode("utf-8"))
        for item in clipboard["items"]:
            if item["label"] != tag:
                continue
            when = parse_ts(item["captured_at"])
            if start <= when <= end:
                events.append((when, decode_fragment(item["encoding"], item["payload"])))

        for line in zf.read("cloud/sync.log").decode("utf-8").splitlines():
            sync_match = re.search(
                r"^\[(?P<ts>[^\]]+)\]\s+tag=(?P<tag>[a-z0-9_-]+)\s+encoding=(?P<enc>\w+)\s+payload=(?P<payload>\S+)$",
                line,
            )
            if not sync_match or sync_match.group("tag") != tag:
                continue
            when = parse_ts(sync_match.group("ts"))
            if start <= when <= end:
                events.append((when, decode_fragment(sync_match.group("enc"), sync_match.group("payload"))))

        drafts_reader = csv.DictReader(zf.read("app/drafts.csv").decode("utf-8").splitlines())
        for row in drafts_reader:
            if row["case_id"] != tag:
                continue
            when = parse_ts(row["saved_at"])
            if start <= when <= end:
                events.append((when, decode_fragment(row["encoding"], row["payload"])))

        for line in zf.read("notes/notes.ndjson").decode("utf-8").splitlines():
            item = json.loads(line)
            if item["tag"] != tag:
                continue
            when = parse_ts(item["saved_at"])
            if start <= when <= end:
                events.append((when, decode_fragment(item["encoding"], item["payload"])))

    events.sort(key=lambda item: item[0])
    flag = "".join(fragment for _, fragment in events)
    print(flag)


if __name__ == "__main__":
    main()
