import base64
import importlib
import re
import sys
from pathlib import Path

import pytest

APP_DIR = Path(__file__).resolve().parent
if str(APP_DIR) not in sys.path:
    sys.path.insert(0, str(APP_DIR))

VAR_PATTERN = re.compile(r'var\s+(_0x[a-f0-9]+)\s*=\s*"([A-Za-z0-9+/=]+)";')
HTML_FLAG_COMMENT_PATTERN = re.compile(r"<!--\s*flag:\s*[^>]+-->", re.IGNORECASE)
JS_LINE_COMMENT_PATTERN = re.compile(r"//\s*([A-Za-z0-9+/=]{8,})")


def decode_twice(value: str) -> str:
    first = base64.b64decode(value).decode("utf-8")
    second = base64.b64decode(first).decode("utf-8")
    return second


@pytest.fixture
def challenge_module(monkeypatch):
    monkeypatch.setenv("CTF_FLAG", "flag{unit_test_flag}")
    import app as challenge_app

    return importlib.reload(challenge_app)


def test_index_embeds_double_base64_runtime_flag_and_no_plaintext(challenge_module):
    client = challenge_module.app.test_client()
    response = client.get("/")
    text = response.get_data(as_text=True)

    assert response.status_code == 200
    assert "flag{unit_test_flag}" not in text
    assert "flag{dev_local_only" not in text

    var_map = {name: value for name, value in VAR_PATTERN.findall(text)}
    assert "_0x4a3f" in var_map
    assert decode_twice(var_map["_0x4a3f"]) == "flag{unit_test_flag}"


def test_index_contains_noise_variables_fake_flags_and_base64_hint(challenge_module):
    text = challenge_module.app.test_client().get("/").get_data(as_text=True)

    var_names = [name for name, _ in VAR_PATTERN.findall(text)]
    assert 6 <= len(var_names) <= 12
    assert len(set(var_names)) == len(var_names)

    fake_flag_comments = HTML_FLAG_COMMENT_PATTERN.findall(text)
    assert len(fake_flag_comments) >= 2

    hints = JS_LINE_COMMENT_PATTERN.findall(text)
    assert hints
    assert any(
        len(base64.b64decode(token + "==", validate=False)) > 0 for token in hints
    )


def test_index_falls_back_to_encoded_dev_placeholder_without_500(monkeypatch):
    monkeypatch.delenv("CTF_FLAG", raising=False)
    import app as challenge_app

    challenge_app = importlib.reload(challenge_app)
    response = challenge_app.app.test_client().get("/")
    text = response.get_data(as_text=True)

    assert response.status_code == 200
    assert "flag{dev_local_only" not in text

    var_map = {name: value for name, value in VAR_PATTERN.findall(text)}
    assert "_0x4a3f" in var_map

    placeholder_flag = getattr(
        challenge_app, "DEV_PLACEHOLDER_FLAG", "flag{dev_local_only_source_audit}"
    )
    assert decode_twice(var_map["_0x4a3f"]) == placeholder_flag
