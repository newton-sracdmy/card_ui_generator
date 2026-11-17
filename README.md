# card-ui

A visual designer for `card_generator` cards. Lay a card out in the browser, check it
against a real Go render, then export the Go file and its test.

This folder is listed in `.git/info/exclude`, so git ignores it without touching
the committed `.gitignore`. Only what you **Export** lands in `card_generator/`.

## Run

```bash
python3 card-ui/server.py            # opens http://127.0.0.1:8765
python3 card-ui/server.py --no-open --port 9000
```

Needs Python 3.9+ (standard library only), the Go toolchain, and the AWS keys in
`.env.dev` / `.env.uat` / `.env` (the same ones `university_card_image_test.go` uses)
to read assets from S3.

## Workflow

1. **Designs → New card.** Pick an output size preset (3942×6174 student/faculty,
   4050×6300, 6337×4101 Saints, …). Tick *different design units* to type
   coordinates in your Figma frame size (e.g. 401×637); they get scaled with
   `sx`, `sy`, `sc` exactly like the hand-written cards.
2. **Add layers.**
   - Drop PNG layers straight onto the canvas: they upload to
     `card-ui/assets/CARD_UI/<design>/…` and get placed. A layer exported at full
     card size is placed edge to edge.
   - Or use **Browse…** on an image to pick from the S3 bucket.
   - Toolbar: Image, Text, Rich text (mixed fonts on one line), Rectangle, Circle,
     QR code (payment or VDS), Photo (member avatar).
3. **Position.** Drag, resize, arrow keys (Shift ×10), snapping (Alt disables it),
   and **↔ Centre on card**.
   - Load the designer's full mockup under *Reference mockup* (shown when nothing
     is selected) and overlay it to line things up.
   - **Y follows** pins a layer below another one, so the underline and labels move
     down when a long name wraps onto two lines.
4. **Text.** Use `{{MemberName}}`, `{{RegNo|card}}`, `{{ExpiryDate|date}}`,
   `{{Department|upper}}`, …
   - Wrap modes: none (`\n` = new line), *balance 2 lines* (`splitNameToLines`),
     *wrap to width*, *shrink to fit*.
   - Vertical gold gradient = `drawVerticalGradientLines`.
5. **Data tab.**
   - *Sample inputs* (employee, membership, merchant, DGePay) drive the preview,
     the Go render and the generated test.
   - *Card fields* decide how `build<Card>CardTemplateData` fills each `data.X`:
     from the member/merchant data with a fallback, or a fixed value.
6. **Render with Go** (Ctrl+Enter). The generated code is compiled into the real
   `card_generator` package with `go test -overlay`, so nothing is written to the
   repo. Every face is then drawn with the real fonts, helpers and S3 loader.
   - Switch between **Go render**, **Overlay** and **Difference** to compare
     with the editor.
   - The editor shows a stand-in QR pattern; the Go render shows the real code.
7. **Export.** Writes:
   - `card_generator/<name>_card_image.go`, containing `is<Card>Card`,
     `Build<Card>CardImages`, `build<Card>CardTemplateData` and one `draw…` per face.
   - `card_generator/cardui_helpers.go`, shared helpers for exported cards.
   - Optionally `Test<Card>CardUIGeneration` in `university_card_image_test.go`.
     It sits between `// card-ui:begin/end` markers, so exporting again replaces it.

   Then:
   - Paste the dispatcher snippet it shows into `BuildUniversityCardImages`.
   - Run `go test ./card_generator -run '^Test<Card>CardUIGeneration$' -v`.
   - Upload any assets Export reports as *only on your machine* to S3 under the
     same key.

## Examples

- `saints_club_example`: the production Saints Club card rebuilt as a design.
  Its generated front renders pixel-identical (0.0002% of pixels differ) to the
  hand-written `drawSaintsClubMemberCardFront`.
- `kitchen_sink`: every element type and option (rotation, tint, rich text,
  attachments, hide-if-empty, wrap modes, rounded corners, Figma units).

## Element → Go

| Layer | Generated Go |
| --- | --- |
| Image | `loadImageFromURIOrURL` + `drawImageContain` / `drawImageCover` / `scaleImageHighQuality`; options `autoCropNonEmpty`, `tintImageToColor` and a fallback fill |
| Text | `drawText` or `drawVerticalGradientText/Lines`; wrapping via `splitNameToLines`, `cardUIWrapLines`, `cardUIFitLines` |
| Rich text | `cardUIDrawSegments` |
| Rectangle / Circle | `DrawRoundedRectangle` / `DrawCircle` with fill and stroke |
| QR code | `loadImageOrGeneratePaymentQRCode` / `loadImageOrGenerateVDSQRCode`, `drawRoundedPanel`, optional centre logo |
| Photo | `drawRectangularAvatar` or `cardUIDrawCircleAvatar` |
| Face corner radius | `cardUIRoundCorners` |

## Files

```
server.py     HTTP server and API (standard library only)
codegen.py    design JSON → Go source, preview harness and test function
build.py      render and export pipelines
gotools.py    go test -c -overlay builds, harness runs, S3 helper
go/           templates for the overlay harnesses and cardui_helpers.go
static/       the editor (plain ES modules, no build step)
designs/      your designs (*.json), autosaved
assets/       local mirror of S3 keys and your uploads; the preview binary runs from here
.build/       compiled harnesses and recent renders
```
# card_ui_generator
