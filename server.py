#!/usr/bin/env python3
"""card-ui — visual designer for card_generator cards.

    python3 card-ui/server.py            # http://127.0.0.1:8765

Design a card in the browser, check it against a real Go render, then export
card_generator/<name>_card_image.go plus a test in university_card_image_test.go.
Standard library only; Go and the repo's .env.* files supply the rest.
"""

import argparse
import hashlib
import json
import mimetypes
import re
import threading
import time
import urllib.request
import webbrowser
from email.utils import formatdate
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import parse_qs, urlparse

import build
import codegen
import gotools
from codegen import CodegenError
from gotools import GoError

UI_DIR = Path(__file__).resolve().parent
STATIC_DIR = UI_DIR / "static"
DESIGN_DIR = UI_DIR / "designs"
ASSET_DIR = gotools.ASSET_DIR
REMOTE_DIR = ASSET_DIR / "_remote"
RENDER_DIR = gotools.BUILD_DIR / "renders"
FONT_DIR = gotools.REPO_DIR / "assets" / "fonts"
FONT_FILES = {
    "AptosRegular": "Aptos.ttf",
    "AptosBold": "Aptos-Bold.ttf",
    "AptosItalic": "Aptos-Italic.ttf",
    "HelveticaRegular": "Helvetica.ttf",
    "HelveticaBold": "Helvetica-Bold.ttf",
    "HelveticaItalic": "Helvetica-Oblique.ttf",
    "HelveticaBoldItalic": "Helvetica-BoldOblique.ttf",
}
SLUG_RE = re.compile(r"^[a-z0-9][a-z0-9_-]{0,80}$")

_fetch_locks = {}
_fetch_locks_guard = threading.Lock()
_list_cache = {}
_font_metrics = None


class HTTPError(Exception):
    def __init__(self, status, message, **extra):
        super().__init__(message)
        self.status = status
        self.extra = extra


# ---------------------------------------------------------------- assets

def asset_path(key):
    key = (key or "").strip().replace("\\", "/")
    if not key or key.startswith("/") or any(part in ("", ".", "..") for part in key.split("/")):
        raise HTTPError(400, f"invalid asset key: {key!r}")
    path = (ASSET_DIR / key).resolve()
    if ASSET_DIR.resolve() not in path.parents:
        raise HTTPError(400, f"invalid asset key: {key!r}")
    return path


def _key_lock(key):
    with _fetch_locks_guard:
        return _fetch_locks.setdefault(key, threading.Lock())


def resolve_asset(src):
    """Return a local file for an S3 key or http(s) URL, downloading it once."""
    src = (src or "").strip()
    if src.startswith(("http://", "https://")):
        ext = Path(urlparse(src).path).suffix.lower()[:6] or ".img"
        path = REMOTE_DIR / (hashlib.sha1(src.encode()).hexdigest() + ext)
        with _key_lock(src):
            if not path.exists():
                req = urllib.request.Request(src, headers={"User-Agent": "card-ui"})
                try:
                    with urllib.request.urlopen(req, timeout=30) as resp:
                        body = resp.read(40 * 1024 * 1024)
                except OSError as exc:
                    raise HTTPError(404, f"could not download {src}: {exc}")
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_bytes(body)
        return path
    path = asset_path(src)
    with _key_lock(src):
        if not path.exists():
            try:
                result = gotools.util("fetch", keys=[src])
            except GoError as exc:
                raise HTTPError(502, f"S3 fetch failed: {exc}", log=exc.log)
            if result.get(src):
                raise HTTPError(404, f"{src}: {result[src]}")
    return path


def list_assets(prefix, refresh=False):
    prefix = prefix.lstrip("/")
    folders, files, error, bucket = set(), {}, None, None
    cached = _list_cache.get(prefix)
    if cached and not refresh and time.time() - cached[0] < 300:
        s3 = cached[1]
    else:
        try:
            s3 = gotools.util("list", prefix=prefix)
            _list_cache[prefix] = (time.time(), s3)
        except GoError as exc:
            s3, error = None, str(exc)
    if s3:
        bucket = s3.get("bucket")
        folders.update(s3.get("folders") or [])
        for f in s3.get("files") or []:
            if not f["key"].endswith("/"):
                files[f["key"]] = {"key": f["key"], "size": f.get("size", 0), "s3": True, "local": False}
    local_dir = ASSET_DIR / prefix if prefix.endswith("/") or not prefix else (ASSET_DIR / prefix).parent
    if local_dir.is_dir() and ASSET_DIR.resolve() in (local_dir.resolve(), *local_dir.resolve().parents):
        for entry in sorted(local_dir.iterdir()):
            rel = entry.relative_to(ASSET_DIR).as_posix()
            if not rel.startswith(prefix) or rel.startswith(("_remote", ".")):
                continue
            if entry.is_dir():
                folders.add(rel + "/")
            elif rel in files:
                files[rel]["local"] = True
            else:
                files[rel] = {"key": rel, "size": entry.stat().st_size, "s3": False, "local": True}
    return {
        "prefix": prefix,
        "bucket": bucket,
        "folders": sorted(folders),
        "files": sorted(files.values(), key=lambda f: f["key"]),
        "error": error,
    }


def font_metrics():
    global _font_metrics
    if _font_metrics is None:
        cache = gotools.BUILD_DIR / "font-metrics.json"
        if cache.exists():
            _font_metrics = json.loads(cache.read_text())
        else:
            _font_metrics = gotools.util("fonts")
            cache.parent.mkdir(parents=True, exist_ok=True)
            cache.write_text(json.dumps(_font_metrics))
    return _font_metrics


# ---------------------------------------------------------------- designs

def design_path(slug):
    if not SLUG_RE.match(slug or ""):
        raise HTTPError(400, f"invalid design name: {slug!r}")
    return DESIGN_DIR / f"{slug}.json"


def list_designs():
    out = []
    for path in sorted(DESIGN_DIR.glob("*.json")):
        try:
            d = json.loads(path.read_text())
        except ValueError:
            continue
        out.append(
            {
                "slug": path.stem,
                "name": d.get("name") or path.stem,
                "faces": [f"{f.get('name')} {f.get('outW')}×{f.get('outH')}" for f in d.get("faces", [])],
                "updated": path.stat().st_mtime,
            }
        )
    return sorted(out, key=lambda d: -d["updated"])


# ---------------------------------------------------------------- HTTP

class Handler(BaseHTTPRequestHandler):
    server_version = "card-ui"

    def log_message(self, fmt, *args):
        if "/api/" in self.path and not self.path.startswith("/api/schema"):
            super().log_message(fmt, *args)

    # -- helpers

    def _json(self, status, payload):
        body = json.dumps(payload).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        self.wfile.write(body)

    def _file(self, path, content_type=None):
        if not path.is_file():
            raise HTTPError(404, "not found")
        stat = path.stat()
        etag = f'"{int(stat.st_mtime * 1000):x}-{stat.st_size:x}"'
        if self.headers.get("If-None-Match") == etag:
            self.send_response(304)
            self.end_headers()
            return
        body = path.read_bytes()
        ctype = content_type or mimetypes.guess_type(path.name)[0] or _sniff(body)
        self.send_response(200)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(body)))
        self.send_header("ETag", etag)
        self.send_header("Last-Modified", formatdate(stat.st_mtime, usegmt=True))
        self.send_header("Cache-Control", "no-cache")
        self.end_headers()
        self.wfile.write(body)

    def _body(self):
        length = int(self.headers.get("Content-Length") or 0)
        return self.rfile.read(length) if length else b""

    def _json_body(self):
        try:
            return json.loads(self._body() or b"{}")
        except ValueError:
            raise HTTPError(400, "invalid JSON body")

    def _dispatch(self, method):
        url = urlparse(self.path)
        query = {k: v[-1] for k, v in parse_qs(url.query).items()}
        try:
            handler = ROUTES.get((method, url.path))
            if handler is None:
                for (m, prefix), fn in PREFIX_ROUTES.items():
                    if m == method and url.path.startswith(prefix):
                        return fn(self, url.path[len(prefix):], query)
                raise HTTPError(404, "not found")
            return handler(self, query)
        except HTTPError as exc:
            self._json(exc.status, {"error": str(exc), **exc.extra})
        except CodegenError as exc:
            self._json(400, {"error": str(exc)})
        except GoError as exc:
            self._json(422, {"error": str(exc), "log": exc.log})
        except (BrokenPipeError, ConnectionResetError):
            pass

    def do_GET(self):
        self._dispatch("GET")

    def do_POST(self):
        self._dispatch("POST")

    def do_PUT(self):
        self._dispatch("PUT")

    # -- routes

    def index(self, query):
        self._file(STATIC_DIR / "index.html", "text/html; charset=utf-8")

    def static(self, rest, query):
        path = (STATIC_DIR / rest).resolve()
        if STATIC_DIR.resolve() not in path.parents:
            raise HTTPError(404, "not found")
        ctype = "text/javascript; charset=utf-8" if path.suffix == ".js" else None
        self._file(path, ctype)

    def font(self, rest, query):
        name = rest.removesuffix(".ttf")
        if name not in FONT_FILES:
            raise HTTPError(404, "unknown font")
        self._file(FONT_DIR / FONT_FILES[name], "font/ttf")

    def render_png(self, rest, query):
        path = (RENDER_DIR / rest).resolve()
        if RENDER_DIR.resolve() not in path.parents:
            raise HTTPError(404, "not found")
        self._file(path, "image/png")

    def api_schema(self, query):
        try:
            metrics, error = font_metrics(), None
        except GoError as exc:
            metrics, error = None, f"{exc}\n{exc.log[-2000:]}"
        self._json(200, dict(codegen.schema(), fontMetrics=metrics, fontMetricsError=error))

    def api_designs(self, query):
        self._json(200, list_designs())

    def get_design(self, slug, query):
        path = design_path(slug)
        if not path.exists():
            raise HTTPError(404, f"no design {slug}")
        self._json(200, json.loads(path.read_text()))

    def put_design(self, slug, query):
        path = design_path(slug)
        if query.get("create") and path.exists():
            raise HTTPError(409, f"a design called {slug} already exists")
        design = self._json_body()
        DESIGN_DIR.mkdir(exist_ok=True)
        tmp = path.with_suffix(".tmp")
        tmp.write_text(json.dumps(design, indent=2))
        tmp.replace(path)
        self._json(200, {"ok": True, "slug": slug, "updated": path.stat().st_mtime})

    def asset(self, query):
        src = query.get("src", "")
        if src.startswith("data:"):
            raise HTTPError(400, "data URIs are used directly")
        self._file(resolve_asset(src))

    def api_assets(self, query):
        self._json(200, list_assets(query.get("prefix", ""), refresh=bool(query.get("refresh"))))

    def api_upload(self, query):
        key = query.get("key", "")
        path = asset_path(key)
        if path.suffix.lower() not in (".png", ".jpg", ".jpeg", ".webp"):
            raise HTTPError(400, "only .png, .jpg, .jpeg and .webp files are supported")
        body = self._body()
        if not body:
            raise HTTPError(400, "empty upload")
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(body)
        _list_cache.clear()
        self._json(200, {"key": key, "size": len(body)})

    def api_codegen(self, query):
        payload = self._json_body()
        gen = build.generate_sources(payload["design"], payload.get("slug"))
        self._json(
            200,
            {
                "fileName": gen["fileName"],
                "code": gen["code"],
                "helpers": gen["helpers"],
                "test": gen["test"],
                "snippet": gen["snippet"],
                "testName": codegen.test_name(payload["design"]),
            },
        )

    def api_render(self, query):
        payload = self._json_body()
        try:
            report = build.render(payload["design"])
        except GoError as exc:
            extra = {}
            try:
                gen = build.generate_sources(payload["design"], payload.get("slug"))
                extra["code"] = gen["code"]
                extra["fileName"] = gen["fileName"]
            except (CodegenError, GoError):
                pass
            raise HTTPError(422, str(exc), log=exc.log, **extra)
        self._json(200, report)

    def api_export(self, query):
        payload = self._json_body()
        slug = payload.get("slug") or ""
        design_path(slug)
        self._json(200, build.export(payload["design"], slug, append_test=bool(payload.get("appendTest", True))))


def _sniff(body):
    if body.startswith(b"\x89PNG"):
        return "image/png"
    if body[:3] == b"\xff\xd8\xff":
        return "image/jpeg"
    if body[:4] == b"RIFF" and body[8:12] == b"WEBP":
        return "image/webp"
    if body.lstrip()[:5] in (b"<svg ", b"<?xml"):
        return "image/svg+xml"
    return "application/octet-stream"


ROUTES = {
    ("GET", "/"): Handler.index,
    ("GET", "/asset"): Handler.asset,
    ("GET", "/api/schema"): Handler.api_schema,
    ("GET", "/api/designs"): Handler.api_designs,
    ("GET", "/api/assets"): Handler.api_assets,
    ("POST", "/api/upload"): Handler.api_upload,
    ("POST", "/api/codegen"): Handler.api_codegen,
    ("POST", "/api/render"): Handler.api_render,
    ("POST", "/api/export"): Handler.api_export,
}
PREFIX_ROUTES = {
    ("GET", "/static/"): Handler.static,
    ("GET", "/fonts/"): Handler.font,
    ("GET", "/renders/"): Handler.render_png,
    ("GET", "/api/designs/"): Handler.get_design,
    ("PUT", "/api/designs/"): Handler.put_design,
}


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--port", type=int, default=8765)
    parser.add_argument("--no-open", action="store_true", help="do not open a browser tab")
    args = parser.parse_args()
    for d in (DESIGN_DIR, ASSET_DIR, gotools.BUILD_DIR):
        d.mkdir(parents=True, exist_ok=True)
    server = ThreadingHTTPServer((args.host, args.port), Handler)
    url = f"http://{args.host}:{args.port}/"
    print(f"card-ui running at {url}  (Ctrl+C to stop)")
    if not args.no_open:
        threading.Timer(0.5, lambda: webbrowser.open(url)).start()
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nbye")


if __name__ == "__main__":
    main()
