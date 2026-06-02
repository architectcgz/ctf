#!/usr/bin/env python3
import json
import os
import subprocess
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from collections import Counter
from datetime import datetime, timezone
from http.cookiejar import CookieJar
from pathlib import Path


def now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


def read_meminfo() -> dict:
    values = {}
    with open("/proc/meminfo", "r", encoding="utf-8") as f:
        for line in f:
            key, raw = line.split(":", 1)
            parts = raw.strip().split()
            if parts:
                values[key] = int(parts[0])
    total = values.get("MemTotal", 0)
    available = values.get("MemAvailable", 0)
    swap_total = values.get("SwapTotal", 0)
    swap_free = values.get("SwapFree", 0)
    return {
        "mem_total_kib": total,
        "mem_available_kib": available,
        "mem_used_kib": max(total - available, 0),
        "swap_total_kib": swap_total,
        "swap_used_kib": max(swap_total - swap_free, 0),
    }


def read_loadavg() -> dict:
    with open("/proc/loadavg", "r", encoding="utf-8") as f:
        one, five, fifteen, *_ = f.read().strip().split()
    return {"load1": float(one), "load5": float(five), "load15": float(fifteen)}


def docker_counts() -> dict:
    try:
        out = subprocess.check_output(
            ["docker", "ps", "--format", "{{.Names}}"],
            text=True,
            timeout=15,
        )
    except Exception as exc:
        return {"error": str(exc)}
    names = [line.strip() for line in out.splitlines() if line.strip()]
    return {
        "docker_ps_total": len(names),
        "ctf_instance_containers": sum(1 for name in names if name.startswith("ctf-instance-")),
        "ctf_workspace_containers": sum(1 for name in names if name.startswith("ctf-workspace-")),
    }


class Session:
    def __init__(self, base_url: str, username: str, password: str):
        self.base_url = base_url.rstrip("/")
        self.username = username
        self.password = password
        self.cookies = CookieJar()
        self.opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(self.cookies))

    def _request(self, method: str, path: str, payload: dict | None = None) -> dict:
        data = None
        headers = {"Accept": "application/json"}
        if payload is not None:
            data = json.dumps(payload).encode("utf-8")
            headers["Content-Type"] = "application/json"
        req = urllib.request.Request(self.base_url + path, data=data, headers=headers, method=method)
        try:
            with self.opener.open(req, timeout=30) as resp:
                body = resp.read().decode("utf-8")
        except urllib.error.HTTPError as exc:
            body = exc.read().decode("utf-8", errors="replace")
            raise RuntimeError(f"{method} {path} http_error={exc.code} body={body}") from exc
        except Exception as exc:
            raise RuntimeError(f"{method} {path} error={exc}") from exc
        try:
            return json.loads(body)
        except Exception as exc:
            raise RuntimeError(f"{method} {path} invalid_json={body}") from exc

    def login(self) -> dict:
        envelope = self._request("POST", "/api/v1/auth/login", {"username": self.username, "password": self.password})
        if envelope.get("code") != 0:
            raise RuntimeError(f"login_failed user={self.username} envelope={envelope}")
        return envelope["data"]

    def create_instance(self, challenge_id: int) -> dict:
        envelope = self._request("POST", f"/api/v1/challenges/{challenge_id}/instances")
        if envelope.get("code") != 0:
            raise RuntimeError(f"create_failed user={self.username} challenge={challenge_id} envelope={envelope}")
        return envelope["data"]

    def list_instances(self) -> list[dict]:
        envelope = self._request("GET", "/api/v1/instances")
        if envelope.get("code") != 0:
            raise RuntimeError(f"list_failed user={self.username} envelope={envelope}")
        return envelope["data"]

    def extend_instance(self, instance_id: int) -> dict:
        envelope = self._request("POST", f"/api/v1/instances/{instance_id}/extend")
        if envelope.get("code") != 0:
            raise RuntimeError(f"extend_failed user={self.username} instance={instance_id} envelope={envelope}")
        return envelope["data"]

    def delete_instance(self, instance_id: int) -> dict:
        envelope = self._request("DELETE", f"/api/v1/instances/{instance_id}")
        if envelope.get("code") != 0:
            raise RuntimeError(f"delete_failed user={self.username} instance={instance_id} envelope={envelope}")
        return envelope


def fetch_health(base_url: str, path: str) -> dict:
    req = urllib.request.Request(base_url.rstrip("/") + path, headers={"Accept": "application/json"}, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            body = resp.read().decode("utf-8")
    except Exception as exc:
        return {"ok": False, "error": str(exc)}
    try:
        return {"ok": True, "payload": json.loads(body)}
    except Exception as exc:
        return {"ok": False, "error": f"invalid_json: {exc}", "body": body}


def write_json(path: Path, payload: dict) -> None:
    path.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def main() -> int:
    base_url = os.environ.get("CTF_BASE_URL", "http://127.0.0.1:8080")
    password = os.environ["CTF_PASSWORD"]
    users = [item.strip() for item in os.environ["CTF_USERS"].split(",") if item.strip()]
    challenges = [int(item.strip()) for item in os.environ["CTF_CHALLENGES"].split(",") if item.strip()]
    run_dir = Path(os.environ["CTF_RUN_DIR"])
    duration_secs = int(os.environ.get("CTF_OBSERVE_DURATION_SECS", "7200"))
    sample_interval_secs = int(os.environ.get("CTF_SAMPLE_INTERVAL_SECS", "300"))
    create_wait_secs = int(os.environ.get("CTF_CREATE_WAIT_SECS", "900"))
    extend_threshold_secs = int(os.environ.get("CTF_EXTEND_THRESHOLD_SECS", "3000"))

    run_dir.mkdir(parents=True, exist_ok=True)
    events_path = run_dir / "events.jsonl"
    status_path = run_dir / "status.json"
    summary_path = run_dir / "summary.json"

    sessions = {username: Session(base_url, username, password) for username in users}
    targets: dict[int, dict] = {}
    extended_ids: set[int] = set()
    errors: list[str] = []
    started_at = time.time()
    monitor_started_at = None

    def append_event(kind: str, payload: dict) -> None:
        event = {"ts": now_iso(), "kind": kind, **payload}
        with events_path.open("a", encoding="utf-8") as f:
            f.write(json.dumps(event, ensure_ascii=False) + "\n")
        write_json(status_path, event)

    def snapshot() -> dict:
        all_rows = {}
        status_counter = Counter()
        expires_at = {}
        remaining_extends = {}
        seen_ids = set()
        for username, session in sessions.items():
            rows = session.list_instances()
            selected = [row for row in rows if row.get("id") in targets]
            all_rows[username] = selected
            for row in selected:
                seen_ids.add(row["id"])
                status_counter[row.get("status", "unknown")] += 1
                expires_at[row["id"]] = row.get("expires_at")
                remaining_extends[row["id"]] = row.get("remaining_extends")
        missing = len(targets) - len(seen_ids)
        if missing > 0:
            status_counter["missing"] += missing
        return {
            "status_counts": dict(status_counter),
            "rows_by_user": all_rows,
            "expires_at": expires_at,
            "remaining_extends": remaining_extends,
            "health": {
                "root": fetch_health(base_url, "/health"),
                "db": fetch_health(base_url, "/health/db"),
                "redis": fetch_health(base_url, "/health/redis"),
            },
            "host": {
                **read_loadavg(),
                **read_meminfo(),
                **docker_counts(),
            },
        }

    try:
        append_event("run_started", {"users": users, "challenges": challenges})
        for username, session in sessions.items():
            session.login()
        append_event("login_ok", {"count": len(sessions)})

        for username, session in sessions.items():
            for challenge_id in challenges:
                data = session.create_instance(challenge_id)
                targets[data["id"]] = {
                    "user": username,
                    "challenge_id": challenge_id,
                    "created_at": data.get("created_at"),
                    "initial_status": data.get("status"),
                }
        write_json(run_dir / "targets.json", {"targets": targets})
        append_event("create_submitted", {"target_count": len(targets), "target_ids": sorted(targets)})

        create_deadline = time.time() + create_wait_secs
        while time.time() < create_deadline:
            snap = snapshot()
            counts = snap["status_counts"]
            append_event("create_poll", {"counts": counts, "host": snap["host"], "health": snap["health"]})
            if counts.get("running", 0) == len(targets):
                monitor_started_at = time.time()
                append_event("create_completed", {"counts": counts})
                break
            bad_count = counts.get("failed", 0) + counts.get("destroying", 0) + counts.get("expired", 0)
            if bad_count > 0:
                raise RuntimeError(f"create_phase_failed counts={counts}")
            time.sleep(10)
        else:
            raise RuntimeError("create_phase_timeout")

        assert monitor_started_at is not None
        end_at = monitor_started_at + duration_secs
        sample_index = 0
        while time.time() < end_at:
            snap = snapshot()
            counts = snap["status_counts"]
            now_ts = datetime.now(timezone.utc)

            for instance_id, expires in snap["expires_at"].items():
                if instance_id in extended_ids or not expires:
                    continue
                try:
                    expire_ts = datetime.fromisoformat(expires.replace("Z", "+00:00"))
                except ValueError:
                    continue
                remaining = (expire_ts - now_ts).total_seconds()
                if remaining <= extend_threshold_secs and snap["remaining_extends"].get(instance_id, 0) > 0:
                    owner = targets[instance_id]["user"]
                    sessions[owner].extend_instance(instance_id)
                    extended_ids.add(instance_id)
                    append_event("instance_extended", {"instance_id": instance_id, "user": owner, "remaining_before_extend_secs": remaining})

            append_event(
                "sample",
                {
                    "index": sample_index,
                    "elapsed_secs": round(time.time() - monitor_started_at, 1),
                    "remaining_secs": round(max(end_at - time.time(), 0), 1),
                    "counts": counts,
                    "health": snap["health"],
                    "host": snap["host"],
                },
            )
            sample_index += 1
            if counts.get("running", 0) != len(targets):
                errors.append(f"non_running_detected counts={counts}")
            sleep_for = min(sample_interval_secs, max(end_at - time.time(), 0))
            if sleep_for > 0:
                time.sleep(sleep_for)

        append_event("cleanup_started", {"target_count": len(targets)})
        target_ids = sorted(targets)
        for index in range(0, len(target_ids), 5):
            batch = target_ids[index : index + 5]
            for instance_id in batch:
                owner = targets[instance_id]["user"]
                sessions[owner].delete_instance(instance_id)
            append_event("cleanup_batch_submitted", {"batch": batch})
            time.sleep(2)

        cleanup_deadline = time.time() + 600
        while time.time() < cleanup_deadline:
            snap = snapshot()
            counts = snap["status_counts"]
            append_event("cleanup_poll", {"counts": counts, "host": snap["host"], "health": snap["health"]})
            terminal = (
                counts.get("stopped", 0)
                + counts.get("destroyed", 0)
                + counts.get("expired", 0)
                + counts.get("failed", 0)
                + counts.get("missing", 0)
            )
            if terminal >= len(targets):
                append_event("cleanup_completed", {"counts": counts})
                break
            time.sleep(10)

        summary = {
            "started_at": datetime.fromtimestamp(started_at, timezone.utc).isoformat(),
            "monitor_started_at": datetime.fromtimestamp(monitor_started_at, timezone.utc).isoformat(),
            "finished_at": now_iso(),
            "duration_secs": duration_secs,
            "sample_interval_secs": sample_interval_secs,
            "users": users,
            "challenges": challenges,
            "target_count": len(targets),
            "extended_instance_ids": sorted(extended_ids),
            "errors": errors,
            "status": "ok" if not errors else "warning",
        }
        write_json(summary_path, summary)
        append_event("run_finished", summary)
        return 0
    except Exception as exc:
        message = str(exc)
        errors.append(message)
        append_event("run_failed", {"error": message})
        write_json(
            summary_path,
            {
                "started_at": datetime.fromtimestamp(started_at, timezone.utc).isoformat(),
                "finished_at": now_iso(),
                "users": users,
                "challenges": challenges,
                "target_count": len(targets),
                "errors": errors,
                "status": "failed",
            },
        )
        return 1


if __name__ == "__main__":
    sys.exit(main())
