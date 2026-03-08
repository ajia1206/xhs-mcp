#!/usr/bin/env python3
"""
Fetch Shanghai Futures Exchange daily datasets for October 2025.

The script hits the documented JSON endpoints directly to avoid parsing
the interactive web UI. Each successful response is saved under
`data/shfe/2025-10/<YYYYMMDD>/<dataset>.json`.

Endpoints can occasionally return 404 on non-trading days; those are logged
and skipped rather than treated as hard failures.
"""

from __future__ import annotations

import datetime as dt
import json
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path
from typing import Dict, Iterable


# Minimal headers keep the CDN happy without pretending to be a browser.
REQUEST_HEADERS: Dict[str, str] = {
    "User-Agent": "Mozilla/5.0 (compatible; shfe-fetch/1.0; +https://example.com)",
    "Accept": "application/json",
}

# Endpoints sourced from https://www.shfe.com.cn/images/api.js
# Only include the daily datasets the UI exposes under “统计数据 > 日周数据”.
DATASETS: Dict[str, str] = {
    "futures_kx": "https://www.shfe.com.cn/data/tradedata/future/dailydata/kx{date}.dat",
    "futures_pm": "https://www.shfe.com.cn/data/tradedata/future/dailydata/pm{date}.dat",
    "futures_timeprice_default": "https://www.shfe.com.cn/data/tradedata/future/dailydata/{date}defaultTimePrice.dat",
    "futures_timeprice_main": "https://www.shfe.com.cn/data/tradedata/future/dailydata/{date}mainTimePrice.dat",
    "futures_timeprice_daily": "https://www.shfe.com.cn/data/tradedata/future/dailydata/{date}dailyTimePrice.dat",
    "futures_markerprice": "https://www.shfe.com.cn/data/tradedata/future/dailydata/{date}markerprice.dat",
    "futures_settlement": "https://www.shfe.com.cn/data/tradedata/future/dailydata/js{date}.dat",
    "futures_dailystock": "https://www.shfe.com.cn/data/tradedata/future/dailydata/{date}dailystock.dat",
    "futures_dlv_premium": "https://www.shfe.com.cn/data/tradedata/future/dailydata/{date}dailydlvplacepremium.dat",
    "options_kx": "https://www.shfe.com.cn/data/tradedata/option/dailydata/kx{date}.dat",
    "options_settlement": "https://www.shfe.com.cn/data/tradedata/option/dailydata/js{date}.dat",
}

# Target date window.
START_DATE = dt.date(2025, 10, 1)
END_DATE = dt.date(2025, 10, 31)

OUTPUT_ROOT = Path("data/shfe/2025-10")


def daterange(start: dt.date, end: dt.date) -> Iterable[dt.date]:
    """Yield each date in the inclusive range [start, end]."""
    delta = dt.timedelta(days=1)
    current = start
    while current <= end:
        yield current
        current += delta


def fetch_json(url: str) -> Dict:
    """Fetch JSON from url and return the parsed payload."""
    # Avoid CDN cache artefacts by appending a lightweight cache buster.
    cache_buster = int(time.time() * 1000)
    sep = "&" if "?" in url else "?"
    req = urllib.request.Request(f"{url}{sep}params={cache_buster}", headers=REQUEST_HEADERS)
    with urllib.request.urlopen(req, timeout=30) as resp:  # nosec B310
        charset = resp.headers.get_content_charset("utf-8")
        data = resp.read().decode(charset)
    try:
        return json.loads(data)
    except json.JSONDecodeError as exc:
        raise ValueError(f"Unexpected payload from {url}") from exc


def main() -> int:
    successes = 0
    skips = 0
    OUTPUT_ROOT.mkdir(parents=True, exist_ok=True)

    for current in daterange(START_DATE, END_DATE):
        stamp = current.strftime("%Y%m%d")
        day_dir = OUTPUT_ROOT / stamp
        day_dir.mkdir(exist_ok=True)

        for name, template in DATASETS.items():
            url = template.format(date=stamp)
            target_file = day_dir / f"{name}.json"

            try:
                payload = fetch_json(url)
            except urllib.error.HTTPError as err:
                if err.code == 404:
                    skips += 1
                    print(f"[skip:{err.code}] {url}", file=sys.stderr)
                    continue
                print(f"[error:{err.code}] {url}", file=sys.stderr)
                return 1
            except Exception as err:
                print(f"[error] {url} -> {err}", file=sys.stderr)
                return 1

            with target_file.open("w", encoding="utf-8") as fh:
                json.dump(payload, fh, ensure_ascii=False)

            successes += 1
            print(f"[ok] {target_file}")

    print(f"Completed: {successes} files saved, {skips} datasets missing (likely non-trading days).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
