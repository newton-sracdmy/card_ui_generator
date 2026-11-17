"""Go toolchain bridge for card-ui.

Generated code is compiled into the real card_generator package with
`go test -c -overlay`, so previews use exactly the same helpers, fonts and
S3 loader as production while nothing is written into the repository.
"""

import hashlib
import json
import os
import shutil
import subprocess
import threading
import time
import uuid
from pathlib import Path

UI_DIR = Path(__file__).resolve().parent

def _find_repo_dir():
    if (UI_DIR / "card_generator").is_dir():
        return UI_DIR
    if "CARD_UI_REPO_DIR" in os.environ:
        return Path(os.environ["CARD_UI_REPO_DIR"]).resolve()
    candidate = UI_DIR.parent
    if (candidate / "card_generator").is_dir():
        return candidate.resolve()
    return UI_DIR

REPO_DIR = _find_repo_dir()
PKG_DIR = REPO_DIR / "card_generator"
ASSET_DIR = UI_DIR / "assets"
BUILD_DIR = UI_DIR / ".build"
TEMPLATE_DIR = UI_DIR / "go"

_build_lock = threading.Lock()
_util_binary = None
_util_hash = None


class GoError(Exception):
    def __init__(self, message, log=""):
        super().__init__(message)
        self.log = log


def _parse_dotenv(path):
    values = {}
    if not path.exists():
        return values
    for raw in path.read_text(errors="replace").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        if line.startswith("export "):
            line = line[len("export "):]
        key, value = line.split("=", 1)
        key, value = key.strip(), value.strip()
        if len(value) >= 2 and value[0] == value[-1] and value[0] in "\"'":
            value = value[1:-1]
        elif " #" in value:
            value = value.split(" #", 1)[0].rstrip()
        values[key] = value
    return values


def go_env():
    """Environment for harness runs, mirroring TestMain's godotenv.Load order."""
    env = {}
    for name in (".env.dev", ".env.uat", ".env"):
        for key, value in _parse_dotenv(REPO_DIR / name).items():
            env.setdefault(key, value)
    env.update(os.environ)
    return env


def gofmt(source, label="generated.go"):
    proc = subprocess.run(["gofmt"], input=source, capture_output=True, text=True)
    if proc.returncode != 0:
        numbered = "\n".join(f"{i + 1:4d}  {line}" for i, line in enumerate(source.splitlines()))
        raise GoError(f"gofmt rejected {label}", proc.stderr + "\n" + numbered)
    return proc.stdout


def _compile(overlay_files, out_binary):
    """Compile the card_generator test binary with virtual files overlaid."""
    work = out_binary.parent
    work.mkdir(parents=True, exist_ok=True)
    replace = {}
    for virtual_name, content in overlay_files.items():
        backing = work / virtual_name
        backing.write_text(content)
        replace[str(PKG_DIR / virtual_name)] = str(backing)
    overlay = work / "overlay.json"
    overlay.write_text(json.dumps({"Replace": replace}, indent=2))
    cmd = ["go", "test", "-c", "-overlay", str(overlay), "-o", str(out_binary), "./card_generator"]
    proc = subprocess.run(cmd, cwd=REPO_DIR, capture_output=True, text=True, timeout=600)
    if proc.returncode != 0:
        raise GoError("Go build failed", proc.stdout + proc.stderr)


def _run(binary, test_name, env_extra, cwd, timeout=600):
    env = go_env()
    env.update(env_extra)
    cmd = [str(binary), "-test.run", f"^{test_name}$", "-test.v", "-test.timeout", f"{timeout}s"]
    proc = subprocess.run(cmd, cwd=cwd, env=env, capture_output=True, text=True, timeout=timeout + 30)
    return proc


def _ensure_util_binary():
    global _util_binary, _util_hash
    source = (TEMPLATE_DIR / "util_test.go.tmpl").read_text()
    digest = hashlib.sha1(source.encode()).hexdigest()
    if _util_binary and _util_binary.exists() and _util_hash == digest:
        return _util_binary
    binary = BUILD_DIR / "util" / "util.test"
    _compile({"zz_cardui_util_test.go": source}, binary)
    _util_binary, _util_hash = binary, digest
    return binary


def util(cmd, **params):
    """Run one TestCardUIUtil command and return its decoded JSON result."""
    with _build_lock:
        binary = _ensure_util_binary()
    out = BUILD_DIR / "util" / f"result-{uuid.uuid4().hex}.json"
    env = {"CARDUI_CMD": cmd, "CARDUI_OUT": str(out), "CARDUI_ASSET_DIR": str(ASSET_DIR)}
    if "keys" in params:
        env["CARDUI_KEYS"] = "\n".join(params["keys"])
    if "prefix" in params:
        env["CARDUI_PREFIX"] = params["prefix"]
    proc = _run(binary, "TestCardUIUtil", env, cwd=BUILD_DIR, timeout=300)
    try:
        result = json.loads(out.read_text())
    except (OSError, ValueError):
        raise GoError(f"util {cmd} failed", proc.stdout + proc.stderr)
    finally:
        out.unlink(missing_ok=True)
    if isinstance(result, dict) and set(result) == {"error"}:
        raise GoError(result["error"], proc.stdout + proc.stderr)
    return result


def render(files, faces):
    """Compile generated sources plus the preview harness and render every face.

    `files` maps virtual file names inside card_generator to Go source. Returns
    the harness report together with the run directory holding <face>.png.
    """
    run_id = time.strftime("%H%M%S-") + uuid.uuid4().hex[:6]
    run_dir = BUILD_DIR / "renders" / run_id
    run_dir.mkdir(parents=True, exist_ok=True)
    binary = BUILD_DIR / "preview" / "preview.test"
    started = time.time()
    with _build_lock:
        _compile(files, binary)
        build_ms = int((time.time() - started) * 1000)
        proc = _run(binary, "TestCardUIPreview", {"CARDUI_OUT": str(run_dir)}, cwd=ASSET_DIR)
    report_path = run_dir / "report.json"
    if not report_path.exists():
        raise GoError("Preview run crashed", proc.stdout + proc.stderr)
    report = json.loads(report_path.read_text())
    report["buildMs"] = build_ms
    report["log"] = proc.stdout + proc.stderr
    report["runId"] = run_id
    _prune_renders(keep=12)
    return report, run_dir


def _prune_renders(keep):
    root = BUILD_DIR / "renders"
    runs = sorted((p for p in root.iterdir() if p.is_dir()), key=lambda p: p.stat().st_mtime)
    for old in runs[:-keep]:
        shutil.rmtree(old, ignore_errors=True)
