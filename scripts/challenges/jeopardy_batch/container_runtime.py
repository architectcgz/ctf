from __future__ import annotations

import re
import socket
import subprocess
import time
import urllib.error
import urllib.request
import uuid
from contextlib import contextmanager
from dataclasses import dataclass
from pathlib import Path
from typing import Iterator

from .paths import IMAGE_TAG, WEB_PORT


CONTAINER_TYPE_RE = re.compile(r"^\s*type:\s*container\s*$", re.M)
SERVICE_PORT_RE = re.compile(r"^\s*port:\s*(\d+)\s*$", re.M)


class ContainerRuntimeError(RuntimeError):
    pass


@dataclass(frozen=True)
class RunningContainer:
    host: str
    port: int
    base_url: str
    service_port: int
    flag: str

    def solver_env(self) -> dict[str, str]:
        return {
            "HOST": self.host,
            "PORT": str(self.port),
            "BASE_URL": self.base_url,
            "SERVICE_PORT": str(self.service_port),
            "FLAG": self.flag,
            "CTF_FLAG": self.flag,
        }


def is_container_pack(pack_dir: Path) -> bool:
    dockerfile = pack_dir / "docker" / "Dockerfile"
    if not dockerfile.exists():
        return False
    challenge_text = (pack_dir / "challenge.yml").read_text(encoding="utf-8")
    return bool(CONTAINER_TYPE_RE.search(challenge_text))


def service_port(pack_dir: Path) -> int:
    challenge_text = (pack_dir / "challenge.yml").read_text(encoding="utf-8")
    match = SERVICE_PORT_RE.search(challenge_text)
    if not match:
        return WEB_PORT
    return int(match.group(1))


def _unique_suffix() -> str:
    return uuid.uuid4().hex[:8]


def _image_name(slug: str) -> str:
    return f"jeopardy-verify-{slug}:{IMAGE_TAG}-{_unique_suffix()}"


def _container_name(slug: str) -> str:
    return f"jeopardy-verify-{slug}-{_unique_suffix()}"


def _run_checked(cmd: list[str], *, timeout: int) -> subprocess.CompletedProcess[str]:
    try:
        return subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=timeout,
            check=True,
        )
    except subprocess.TimeoutExpired as exc:
        output = ((exc.stdout or "") + "\n" + (exc.stderr or "")).strip()
        raise ContainerRuntimeError(f"命令超时: {' '.join(cmd)}\n{output}".strip()) from exc
    except subprocess.CalledProcessError as exc:
        output = ((exc.stdout or "") + "\n" + (exc.stderr or "")).strip()
        raise ContainerRuntimeError(f"命令失败: {' '.join(cmd)}\n{output}".strip()) from exc


def _docker_logs(container_name: str) -> str:
    proc = subprocess.run(
        ["docker", "logs", container_name],
        capture_output=True,
        text=True,
        timeout=15,
    )
    return ((proc.stdout or "") + "\n" + (proc.stderr or "")).strip()


def _host_port(container_name: str, container_port: int) -> int:
    proc = _run_checked(
        ["docker", "port", container_name, f"{container_port}/tcp"],
        timeout=15,
    )
    match = re.search(r":(\d+)\s*$", proc.stdout.strip())
    if not match:
        raise ContainerRuntimeError(
            f"无法解析容器端口映射: {container_name} {container_port}/tcp -> {proc.stdout.strip()}"
        )
    return int(match.group(1))


def _wait_for_tcp(host: str, port: int, *, timeout_seconds: float = 20.0) -> None:
    deadline = time.monotonic() + timeout_seconds
    while time.monotonic() < deadline:
        try:
            with socket.create_connection((host, port), timeout=0.5):
                return
        except OSError:
            time.sleep(0.25)
    raise ContainerRuntimeError(f"服务未就绪: {host}:{port}")


def _wait_for_http(base_url: str, *, timeout_seconds: float = 20.0) -> None:
    deadline = time.monotonic() + timeout_seconds
    while time.monotonic() < deadline:
        try:
            with urllib.request.urlopen(base_url, timeout=1.0) as response:
                response.read(1)
            return
        except urllib.error.HTTPError:
            return
        except (urllib.error.URLError, ConnectionError, OSError):
            time.sleep(0.25)
    raise ContainerRuntimeError(f"HTTP 服务未就绪: {base_url}")


def _build_image(pack_dir: Path, image_name: str) -> None:
    last_error: ContainerRuntimeError | None = None
    for attempt in range(1, 4):
        try:
            _run_checked(
                ["docker", "build", "-t", image_name, str(pack_dir / "docker")],
                timeout=600,
            )
            return
        except ContainerRuntimeError as exc:
            last_error = exc
            if attempt == 3:
                break
            time.sleep(attempt * 2)
    assert last_error is not None
    raise last_error


def _run_container(image_name: str, container_name: str, *, container_port: int, flag_value: str) -> str:
    proc = _run_checked(
        [
            "docker",
            "run",
            "-d",
            "--name",
            container_name,
            "-p",
            f"127.0.0.1::{container_port}",
            "-e",
            f"FLAG={flag_value}",
            "-e",
            f"CTF_FLAG={flag_value}",
            "-e",
            f"PORT={container_port}",
            image_name,
        ],
        timeout=30,
    )
    return proc.stdout.strip()


def _cleanup_container(container_name: str) -> None:
    subprocess.run(
        ["docker", "rm", "-f", container_name],
        capture_output=True,
        text=True,
        timeout=30,
    )


def _cleanup_image(image_name: str) -> None:
    subprocess.run(
        ["docker", "image", "rm", "-f", image_name],
        capture_output=True,
        text=True,
        timeout=30,
    )


@contextmanager
def running_verification_container(
    pack_dir: Path,
    slug: str,
    flag_value: str,
    *,
    wait_mode: str = "http",
) -> Iterator[RunningContainer]:
    if not is_container_pack(pack_dir):
        raise ContainerRuntimeError(f"题包不是容器题: {pack_dir}")

    image_name = _image_name(slug)
    container_name = _container_name(slug)
    internal_port = service_port(pack_dir)
    container_started = False

    try:
        _build_image(pack_dir, image_name)
        _run_container(
            image_name,
            container_name,
            container_port=internal_port,
            flag_value=flag_value,
        )
        container_started = True
        host = "127.0.0.1"
        port = _host_port(container_name, internal_port)
        base_url = f"http://{host}:{port}"
        try:
            if wait_mode == "tcp":
                _wait_for_tcp(host, port)
            else:
                _wait_for_http(base_url)
        except ContainerRuntimeError as exc:
            logs = _docker_logs(container_name)
            raise ContainerRuntimeError(f"{exc}\n容器日志:\n{logs}".strip()) from exc
        yield RunningContainer(
            host=host,
            port=port,
            base_url=base_url,
            service_port=internal_port,
            flag=flag_value,
        )
    finally:
        if container_started:
            _cleanup_container(container_name)
        _cleanup_image(image_name)
