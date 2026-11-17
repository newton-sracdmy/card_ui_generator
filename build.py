"""Render and export pipelines shared by the HTTP server and the CLI."""

import re

import codegen
import gotools
from codegen import CodegenError, PKG_DIR, TEST_FILE, HELPERS_FILE


def generate_sources(design, slug=None):
    label = f"card-ui/designs/{slug}.json" if slug else None
    gen = codegen.generate(design, source_label=label)
    codegen.check_target_file(gen["fileName"])
    collisions = codegen.find_collisions(gen["declared"], gen["fileName"])
    if collisions:
        raise CodegenError(
            "Generated names clash with existing card_generator code: "
            + "; ".join(collisions)
            + ". Change the card name or the Go identifier in Settings."
        )
    gen["code"] = gotools.gofmt(gen["code"], gen["fileName"])
    gen["helpers"] = gotools.gofmt(codegen.helpers_source(), HELPERS_FILE)
    gen["test"] = _format_test_block(codegen.gen_test_function(design, gen))
    gen["snippet"] = codegen.dispatcher_snippet(design)
    return gen


def _format_test_block(block):
    formatted = gotools.gofmt("package card_generator\n\n" + block, "test function")
    return formatted.split("\n", 2)[2].strip("\n") + "\n"


def render(design):
    gen = generate_sources(design)
    harness = gotools.gofmt(codegen.gen_preview_harness(design, gen), "preview harness")
    files = {
        gen["fileName"]: gen["code"],
        HELPERS_FILE: gen["helpers"],
        "zz_cardui_preview_test.go": harness,
    }
    report, run_dir = gotools.render(files, gen["faces"])
    faces = {}
    for face_id, _ in gen["faces"]:
        if (run_dir / f"{face_id}.png").exists():
            faces[face_id] = f"/renders/{run_dir.name}/{face_id}.png"
    report["faces"] = faces
    return report


def export(design, slug, append_test=True):
    gen = generate_sources(design, slug)
    written = []
    (PKG_DIR / gen["fileName"]).write_text(gen["code"])
    written.append(f"card_generator/{gen['fileName']}")
    helpers_path = PKG_DIR / HELPERS_FILE
    if not helpers_path.exists() or helpers_path.read_text() != gen["helpers"]:
        helpers_path.write_text(gen["helpers"])
        written.append(f"card_generator/{HELPERS_FILE}")
    if append_test:
        _upsert_test(gen["test"], codegen.test_name(design))
        written.append(f"card_generator/{TEST_FILE.name} ({codegen.test_name(design)})")

    warnings = []
    s3_keys = [k for k in gen["assets"] if not k.startswith(("{{",))]
    if s3_keys:
        try:
            exists = gotools.util("exists", keys=s3_keys)
            not_on_s3 = [k for k in s3_keys if not exists.get(k)]
            local_only = [k for k in not_on_s3 if (gotools.ASSET_DIR / k).is_file()]
            nowhere = [k for k in not_on_s3 if k not in local_only]
            if local_only:
                warnings.append(
                    "These assets are only on your machine (card-ui/assets) — upload them to S3 under the same key "
                    "before deploying: " + ", ".join(local_only)
                )
            if nowhere:
                warnings.append(
                    "These assets were not found on S3 or locally, so Go will draw the fallback (or nothing): "
                    + ", ".join(nowhere)
                )
        except gotools.GoError as exc:
            warnings.append(f"Could not verify assets on S3: {exc}")
    return {
        "written": written,
        "snippet": gen["snippet"],
        "testName": codegen.test_name(design),
        "warnings": warnings,
        "runCommand": f"go test ./card_generator -run '^{codegen.test_name(design)}$' -v",
    }


def _upsert_test(block, name):
    src = TEST_FILE.read_text()
    pattern = re.compile(
        rf"// card-ui:begin {re.escape(name)}\n.*?// card-ui:end {re.escape(name)}\n?", re.S
    )
    if pattern.search(src):
        src = pattern.sub(lambda _: block, src, count=1)
    else:
        if re.search(rf"^func {re.escape(name)}\(", src, re.M):
            raise CodegenError(
                f"{TEST_FILE.name} already has a hand-written {name}; rename the design or remove that test first."
            )
        src = src.rstrip("\n") + "\n\n" + block
    TEST_FILE.write_text(src)
