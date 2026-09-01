<div align="center">

# 🧠 GLM-Free-API

**یک فایل اجرا کن — یک API کامل داشته باش.**

[![Go](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Release](https://img.shields.io/badge/Release-download--one--file-8A2BE2?logo=github)](https://github.com/Godde3s/glm-free-api/releases)
[![Multi-Account](https://img.shields.io/badge/Multi--Account-round--robin%20%2B%20failover-FF6B35)](#-چند-اکانت-بدون-لیمیت)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**پل مستقیم به مدل‌های GLM** روی [chat.z.ai](https://chat.z.ai) — بدون مرورگر، بدون درایور، بدون دردسر.
هر ابزاری که با API هوش مصنوعی کار کند، بلافاصله به GLM وصل می‌شود.

`OpenAI /v1/chat/completions` · `Anthropic /v1/messages` · استریم واقعی SSE · تصویر (Vision) · ابزار (Tool-Calling) · **پشتیبانی چند اکانت** · [داشبورد وب فارسی](#-داشبورد-وب--قلب-برنامه)

</div>

---

<div dir="rtl">

# 🇮🇷 راهنمای فارسی — از صفر تا API در ۵ دقیقه

## این چیز اصلاً چیست؟

فرض کنید یک **مترجم جادویی** دارید: هر برنامه‌ای که فقط زبانِ «OpenAI API» را بلد است، با این مترجم حرف می‌زند و پشت صحنه، حرف‌هایش به GLM (مدل هوشمند Z.AI) منتقل و جوابش برمی‌گردد. شما هیچ تفاوتی حس نمی‌کنید — انگار یک API واقعی و پولی در اختیار دارید، ولی رایگان.

این پل **بدون مرورگر** کار می‌کند (خلاف اکثر پروژه‌های مشابه). یعنی نه کروم لازم دارد، نه درایور گرافیکی، نه سرور قوی — یک فایل باینری کوچک که روی هر سیستمی اجرا می‌شود.

## 🚀 شروع سریع — فقط یک مرحله

### راه ۱: دانلود فایل آماده (ساده‌ترین — بدون نصب هیچ‌چیز)

1. برو به صفحه [**Releases**](https://github.com/Godde3s/glm-free-api/releases)
2. فایل مخصوص سیستم‌عاملت را دانلود و باز کن:
   - ویندوز → `zai-api-windows-amd64.zip`
   - مک (جدید) → `zai-api-darwin-arm64.tar.gz`
   - لینوکس → `zai-api-linux-amd64.tar.gz`
3. **ویندوز:** `start.bat` را دابل‌کلیک کن — تمام! ✅
   **مک/لینوکس:** در ترمینال همان پوشه بزن: `./start.sh` — تمام! ✅

سرور روشن می‌شود و آدرس‌ها را نشان می‌دهد. همین.

> 💡 بعد از اجرا، داشبورد فارسی را باز کن: **http://localhost:3001** — تست زنده، سلامت اکانت‌ها و کد آماده اتصال، همه در یک صفحه.

### راه ۲: با داکر (برای سرور)

```bash
# توکن‌هایت را در .env بگذار (کپی از .env.example) بعد:
docker compose up -d
```

### راه ۳: خودت بیلد کن (برای برنامه‌نویس‌ها)

```bash
git clone https://github.com/Godde3s/glm-free-api.git && cd glm-free-api
./start.sh          # خودش بیلد می‌کند و اجرا می‌کند
```

## 🖥️ داشبورد وب — قلب برنامه

وقتی سرور روشن شد، مرورگر را باز کن:

<div align="center">

**http://localhost:3001**

</div>

یک پنل کنترل فارسی و تیره (داخل خود فایل — بدون هیچ فایل اضافه) که همین‌جا همه‌چیز را می‌بینی:

- 🟢 **وضعیت زنده** سرور و تک‌تک اکانت‌ها (سالم / در حال استراحت با شمارش معکوس / مُرد)
- 🔌 **پنل اتصال**: Base URL و API Key با دکمه کپی — برای وصل‌کردن Cursor، Cline یا هر ابزار دیگه
- 🧪 **تست زنده**: مدل را انتخاب کن، پیام بفرست، استریم واقعی جواب را همان‌جا ببین — یعنی مطمئن می‌شی همه‌چیز کار می‌کنه بدون حتی ترمینال باز کردن
- 📋 **کد آماده**: curl / Python / Node.js / تنظیم ایجنت — با مقادیر واقعی خودت، آماده کپی

## 🔑 توکن رایگان از کجا بیاورم؟ (۱۰ ثانیه)

> بدون توکن هم می‌شود! (حالت مهمان فقط مدل `x-preview-l` را می‌دهد)
> ولی با توکن، همه مدل‌ها + ارسال عکس + چند اکانت فعال می‌شود.
>
> ⚠️ **برای چت، در هر دو حالت `tokens.sqlite` باید دستگاه‌توکن داشته باشد** (کپچای Z.AI).
> یک بار `./token-collector` را اجرا کن — بقیه‌اش خودکار است.

1. برو به **[chat.z.ai](https://chat.z.ai)** و لاگین کن (حساب گوگل، چند ثانیه)
2. دکمه **F12** را بزن → تب **Console** را باز کن
3. این خط را کپی کن، پیست کن و Enter بزن:

```js
localStorage.getItem('token')
```

4. یک متن طولانی (JWT) ظاهر می‌شود → روی خروجی **راست‌کلیک → Copy string contents**

تمام! حالا در فایل `.env` کنار برنامه (اگر نیست، از `.env.example` کپی کن) این را بنویس:

```ini
ZAI_TOKENS=توکن-کپی‌شده
```

<div align="center">

**💡 چند اکانت داری؟ با کاما جدا کن — لیimit برای همیشه مُرد!**

```ini
ZAI_TOKENS=توکن۱,توکن۲,توکن۳
```

</div>

## ⚡ قابلیت چند اکانت — دور زدن لیمیت

وقتی چند توکن بدهی، این اتفاق‌ها می‌افتد (خودکار، بدون هیچ تنظیمی):

- **نوبت‌دهی هوشمند (Round-Robin):** هر درخواست به اکانت بعدی می‌رود؛ فشار بین همه پخش می‌شود
- **فیل‌اُوِر شفاف:** اگر اکانتی لیمیت بخورد (429)، درخواست *قبل از رسیدن حتی یک بایت به برنامه‌ات* به اکانت سالم می‌پرد — برنامه‌ات هیچ خطایی نمی‌بیند!
- **استراحت هوشمند:** اکانت لیمیت‌خورده مدتی استراحت می‌کند (شروع ۶۰ ثانیه، هر بار ×۲، سقف ۳۰ دقیقه) و خودش برمی‌گردد
- **تشخیص مرگ توکن:** توکن منقضی (401) خودکار از گردش خارج می‌شود
- **مانیتورینگ زنده:** آدرس `/status` را باز کن — سلامت تک‌تک اکانت‌ها را می‌بینی

## 📡 حالا چطور استفاده کنم؟

هر جا که «OpenAI API» می‌خواهد، این را بده:

| تنظیم | مقدار |
|---|---|
| **Base URL** | `http://localhost:3001/v1` |
| **API Key** | `Waguri` (یا هر چیزی که در `.env` گذاشتی) |
| **مدل‌ها** | `x-preview-l` (مهمان) · `glm-5.3` · `glm-5.2` · `GLM-5v-Turbo` و... (با توکن) |

**نمونه با curl:**

```bash
curl http://localhost:3001/v1/chat/completions \
  -H "Authorization: Bearer Waguri" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "x-preview-l",
    "messages": [{"role": "user", "content": "سلام! خودت رو معرفی کن"}],
    "stream": true
  }'
```

**در پایتون با کتابخانه رسمی OpenAI:**

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:3001/v1",
    api_key="Waguri",
)

resp = client.chat.completions.create(
    model="glm-5.3",
    messages=[{"role": "user", "content": "برنامه‌ریزی سفر ۳ روزه به استانبول"}],
)
print(resp.choices[0].message.content)
```

**برای ایجنت‌ها (Cursor، Cline، خودت AI و...):** در تنظیمات هر ابزار، فقط Base URL و API Key را همین بگذار و `AGENT_MODE=1` را در `.env` فعال کن — ابزارها (Tool-Calling) هم پشتیبانی می‌شوند.

## 🩺 عیب‌یابی سریع

| مشکل | علت و راه‌حل |
|---|---|
| خطای `401` در چت | توکن منقضی شده — توکن تازه بگیر و `.env` را آپدیت کن |
| `429 / rate_limit` زیاد | اکانت‌های بیشتر اضافه کن: `ZAI_TOKENS=t1,t2,t3` |
| عکس قبول نمی‌کند | حتماً توکن گذاشته باشی + مدل vision مثل `GLM-5v-Turbo` |
| پورت اشغال است | در `.env` بگذار: `PORT=3002` |
| ابزارها (tools) کار نمی‌کند | `.env` → `AGENT_MODE=1` و ری‌استارت |

</div>

---

<div align="center">

# 🇬🇧 English Documentation

</div>

## 🖥️ Web dashboard

Open **http://localhost:3001** after starting — a self-contained Persian/RTL dark dashboard compiled into the binary (no CDN, no assets, works offline):

- live server + **per-account health** (healthy / cooling-down countdown / dead)
- connect panel: base URL + key with copy buttons
- **streaming playground** — pick a model, send a prompt, watch real SSE output
- ready snippets: curl / Python / Node / agent setup, pre-filled with your URLs

## What is this?

An **OpenAI- and Anthropic-compatible API bridge** for [chat.z.ai](https://chat.z.ai) — written in pure Go, **zero browser automation at runtime** (HTTP/2 directly, like the real web client). Point any AI tool at it and it behaves exactly like a real GLM API: streaming SSE, vision, tool-calling, thinking modes, live model catalog.

Forked and heavily upgraded from [izaart95-jpg/GLM-Free-API](https://github.com/izaart95-jpg/GLM-Free-API).

## ✨ What this fork adds

| | Original | **This fork** |
|---|---|---|
| Accounts | 1 token (or guest) | **Multi-account pool** — round-robin + transparent failover |
| 429 rate limit | request fails | **before-stream failover** to a healthy account, exponential cooldown |
| First run | collector + DB + go build required | **one file runs it** (`.env` optional, DB auto-created) |
| Config | env vars only | **`.env` file support** (real env still wins) |
| Install | clone + build | **prebuilt binaries** for 5 platforms via Releases |
| First contact | curl needed | **built-in web dashboard** with live test playground |
| Observability | session status | **per-account health** in `/health` + `/status` + dashboard |
| Tests | integration only | **+8 unit tests** for the pool (cooldown, failover, binding...) |

## 🚀 Quick start

**Option A — download a binary** from [Releases](https://github.com/Godde3s/glm-free-api/releases), unzip, run `./start.sh` (macOS/Linux) or double-click `start.bat` (Windows).

**Option B — Docker:**

```bash
cp .env.example .env   # paste your ZAI_TOKENS
docker compose up -d
```

**Option C — from source:**

```bash
git clone https://github.com/Godde3s/glm-free-api.git && cd glm-free-api
./start.sh   # auto-builds with Go, then runs
```

### Getting a `ZAI_TOKEN`

1. Log in at **https://chat.z.ai**
2. `F12` → Console → `localStorage.getItem('token')`
3. Copy the JWT → paste into `.env`:

```ini
ZAI_TOKENS=token1,token2,token3   # multi-account: commas
AUTH_TOKEN=Waguri                 # your clients' key (change if exposed!)
AGENT_MODE=1                      # for agent frameworks (tool-calling)
```

No tokens = guest mode (`x-preview-l` works, vision does not).

> **Captcha note (both modes):** chat completions need **device tokens** in
> `tokens.sqlite` — they power the captcha Z.AI requires. Seed once with
> `./token-collector` (build: `go build -o token-collector ./cmd/token-collector`).
> Each chat consumes one device token; 200 tokens ≈ 200 replies, re-run anytime.
> Without them the API returns a clear `FRONTEND_CAPTCHA_REQUIRED` error.

## 🔀 Multi-account pool

Give `ZAI_TOKENS=a,b,c` and the bridge becomes a self-healing cluster:

- **Round-robin** — each request takes the next healthy account
- **Pre-stream failover** — a 429/401/403 on one account transparently retries the *same* request on the next account before a single byte reaches your client (nothing that already streamed is ever retried)
- **Exponential cooldown** — `ACCOUNT_COOLDOWN_BASE` (60s default) doubles per consecutive 429, capped at 30 min; successes reset the streak
- **Dead detection** — 401/403 marks the token expired and removes it from rotation
- **Bounded queue** — when *every* account is cooling down, requests wait up to `ACCOUNT_QUEUE_TIMEOUT` (120s) instead of erroring instantly
- **Owner-bound cleanup** — throwaway chats are deleted with the token that created them (no orphaned sessions)
- **Live telemetry** — `/status` shows per-account snapshots; `/health` is pool-aware

Tuning: `ACCOUNT_COOLDOWN_BASE`, `ACCOUNT_QUEUE_TIMEOUT` (see `.env.example`).

## Features (inherited + hardened)

- **Dual protocol** — OpenAI `/v1/chat/completions` + Anthropic `/v1/messages`; streaming (SSE, 5 s keep-alives) and non-streaming
- **Vision** — images on both endpoints (OpenAI `image_url`, Anthropic `image` blocks; URL or base64), up to **10 images / 50 MB each**; uploads ride the *serving account's* token
- **Agent mode** — translates OpenAI tools/roles into Z.AI's prompt format and rewrites `<<<TOOL_CALL>>>` blocks back into native `tool_calls` / `tool_use` (modern XML shim by default)
- **Throwaway sessions** — every chat is deleted on Z.AI the moment its response completes (no context rot, no dead-session buildup); pre-warmed session pool by default (`--sync-mode` for the legacy flow)
- **Live model catalog** — fetched from Z.AI `/api/models` (5-min cache, static fallback), with OpenAI-style `architecture` metadata
- **Per-model features** — `reasoning_effort` (`high`/`max`) forwarded when supported; `image_generation` always off
- **Pure HTTP/2** — no Playwright/Selenium/CGO; static single binary
- **Graceful shutdown** — CTRL+C drains requests, deletes pooled chats, exits clean

## Configuration

Environment variables (or `.env` — same names, real env wins):

| Variable | Default | Description |
|---|---|---|
| `ZAI_TOKENS` | *(empty)* | **Multi-account pool**, comma/semicolon/newline separated |
| `ZAI_TOKEN` | *(empty)* | Legacy single token (`ZAI_TOKENS` wins) |
| `PORT` / `HOST` | `3001` / `0.0.0.0` | Listen address |
| `AUTH_TOKEN` | `Waguri` | Key clients must send (Bearer / `x-api-key`) |
| `ACCOUNT_COOLDOWN_BASE` | `60` | First 429 cooldown (seconds), doubles per streak |
| `ACCOUNT_QUEUE_TIMEOUT` | `120` | Seconds to wait when all accounts are cooling (0 = forever) |
| `AGENT_MODE` | `false` | Tool-calling translation shim |
| `AGENT_MODE_VARIANT` | `modern` | `modern` (XML sections) or `legacy` ([ROLE: ...]) |
| `SESSION_POOL_SIZE` | `5` | Pre-made chat sessions kept ready |
| `SESSION_ACQUIRE_TIMEOUT` | `10` | Seconds to wait for a pooled session |
| `STREAM_HOLDBACK` | `24` | Runes held back to absorb `edit_content` backtracks (0 = off) |
| `SYNC_MODE` | `false` | Fresh chat per request (no pre-warm) |

CLI flags: `--db-path`, `--verbose`, `--agent-mode`, `--agent-mode-variant`, `--sync-mode`.

## API reference

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/v1/chat/completions` | ✅ | OpenAI chat (stream + non-stream) |
| `POST` | `/v1/messages` | ✅ | Anthropic Messages (stream + non-stream, tool use, thinking) |
| `GET` | `/v1/models` | ✅ | OpenAI-style model list |
| `GET` | `/health` | ❌ | Health incl. account pool summary |
| `GET` | `/status` | ❌ | Session + **per-account snapshots** |
| `POST` | `/features` | ✅ | Per-model feature overrides |
| `GET` | `/features` | ✅ | Inspect resolved features |

Request fields beyond the OpenAI standard: `reasoning` / `thinking` (enable thinking), `reasoning_effort` (`high`/`max`), `webSearch`/`search` (toggle web search), `tools` (agent mode).

### Vision example

```json
{"model": "GLM-5v-Turbo",
 "messages": [{"role": "user", "content": [
   {"type": "text", "text": "What is in this image?"},
   {"type": "image_url", "image_url": {"url": "data:image/png;base64,iVBORw..."}}
]}]}
```

## Building & testing

```bash
go build -trimpath -ldflags="-s -w" -o zai-api .   # static binary
go test ./...                                      # unit + integration
```

Releases are cut by tagging: `git tag v1.0.1 && git push --tags` — the workflow builds static binaries for linux (amd64/arm64), macOS (intel/arm) and Windows and attaches them to the release.

## Project structure

```
main.go                    thin entry point
internal/zbridge/          the bridge: accounts, session pool, captcha,
                           agent shim, dual-protocol handlers, vision, SSE
cmd/token-collector/       optional: harvest device tokens for guest/captcha
image-gen/                 helper script for Z.AI image generation
.github/workflows/         release pipeline
```

## FAQ

**Is this really API quality?** Streaming, tool-calls, vision and thinking behave like the official APIs; the upstream is the same web endpoint the chat UI uses. Sessions are throwaway, so no server-side history pollutes your context.

**Where do my conversations live?** Nowhere. Each request is stateless; the chat is deleted on Z.AI after the response. Your state lives entirely in your client's `messages` array.

**Guest mode?** Works for quick tests (model `x-preview-l` only — everything else answers 403 `user level`), rate limits are tighter and vision is unavailable. One ZAI_TOKEN fixes all of that; several tokens make it a cluster.

**My token expired — how do I know?** Its snapshot shows `"dead": true` in `/status`; grab a fresh one from chat.z.ai and update `.env`.

## Security notes

- Set a **strong `AUTH_TOKEN`** before exposing the bridge beyond localhost — anyone with the key uses your accounts.
- Tokens in `.env` are live credentials; don't commit it (`.gitignore` already excludes it) and don't paste them in chats/screenshots.
- This project bridges your own Z.AI accounts. Respect Z.AI's terms; heavy abuse gets tokens banned (the pool dead-detection will tell you when).

## Credits

Built on [izaart95-jpg/GLM-Free-API](https://github.com/izaart95-jpg/GLM-Free-API) (MIT), which itself follows the architecture of [DeepseekFreeAPI](https://github.com/izaart95-jpg/DeepseekFreeAPI). Multi-account pool, one-file UX, release pipeline and Persian docs by [Godde3s](https://github.com/Godde3s).

## License

MIT — see [LICENSE](LICENSE).
