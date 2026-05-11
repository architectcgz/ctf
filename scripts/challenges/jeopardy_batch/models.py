from __future__ import annotations

from dataclasses import dataclass, field
from typing import Callable


@dataclass(frozen=True)
class Target:
    slug: str
    title: str
    category: str
    mode: str
    builder: str
    kind: str
    difficulty: str
    points: int
    primary_skill: str
    primary_action: str
    training_goal: str

@dataclass
class BuildResult:
    statement: str
    solution: str
    solve_py: str
    hints: list[str]
    files: dict[str, bytes] = field(default_factory=dict)
    attachments: list[tuple[str, str]] = field(default_factory=list)
    flag_type: str = "static"
    flag_value: str | None = None
    runtime_type: str = "none"
    runtime_port: int | None = None


Builder = Callable[[Target], BuildResult]
