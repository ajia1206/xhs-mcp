#!/usr/bin/env python3
"""
Normalize SHFE daily JSON snapshots (October 2025) into tabular CSV files.

The script scans `data/shfe/2025-10/<YYYYMMDD>` for the JSON artifacts created
by `fetch_shfe_oct2025.py`, flattens each dataset into one row per record, and
stores monthly aggregates under `data/shfe/2025-10/structured/<dataset>.csv`.

Only shallow key/value pairs are preserved – the source payloads already use
flat dictionaries. A `trade_date` column is appended to help downstream joins.
"""

from __future__ import annotations

import csv
import json
from collections import defaultdict
from io import StringIO
from pathlib import Path
from typing import Dict, Iterable, List, Sequence

INPUT_ROOT = Path("data/shfe/2025-10")
OUTPUT_ROOT = INPUT_ROOT / "structured"
WEB_STRUCT_ROOT = Path("web/data/structured")

# dataset -> JSON list key to extract
DATASET_SPECS: Dict[str, str] = {
    "futures_kx": "o_curinstrument",
    "futures_pm": "o_cursor",
    "futures_timeprice_default": "o_currefprice",
    "futures_timeprice_main": "o_currefprice",
    "futures_timeprice_daily": "o_currefprice",
    "futures_markerprice": "o_curMarkerPrice",
    "futures_settlement": "o_cursor",
    "futures_dailystock": "o_cursor",
    "futures_dlv_premium": "o_cursor",
    "options_kx": "o_curinstrument",
    "options_settlement": "o_cursor",
}


def iter_day_dirs(root: Path) -> Iterable[Path]:
    """Yield child directories whose names look like YYYYMMDD."""
    for child in sorted(root.iterdir()):
        if child.is_dir() and len(child.name) == 8 and child.name.isdigit():
            yield child


def load_records(path: Path, list_key: str, trade_date: str) -> List[Dict]:
    """Read JSON file and return flattened records with trade_date column."""
    payload = json.loads(path.read_text(encoding="utf-8"))
    records = payload.get(list_key, [])
    if not isinstance(records, list):
        return []
    output: List[Dict] = []
    for item in records:
        if isinstance(item, dict):
            row = dict(item)
            row["trade_date"] = trade_date
            output.append(row)
    return output


def collect_dataset(dataset: str, list_key: str) -> List[Dict]:
    """Aggregate records for a single dataset across all trading days."""
    rows: List[Dict] = []
    for day_dir in iter_day_dirs(INPUT_ROOT):
        trade_date = day_dir.name
        src = day_dir / f"{dataset}.json"
        if not src.exists():
            continue
        rows.extend(load_records(src, list_key, trade_date))
    return rows


def render_csv(rows: Sequence[Dict]) -> str:
    """Render rows to CSV string using union of keys as header."""
    if not rows:
        return ""
    fieldnames: List[str] = []
    seen = set()
    for row in rows:
        for key in row.keys():
            if key not in seen:
                seen.add(key)
                fieldnames.append(key)
    buffer = StringIO()
    writer = csv.DictWriter(buffer, fieldnames=fieldnames)
    writer.writeheader()
    for row in rows:
        writer.writerow(row)
    return buffer.getvalue()


def main() -> int:
    if not INPUT_ROOT.exists():
        raise SystemExit(f"Input directory not found: {INPUT_ROOT}")
    OUTPUT_ROOT.mkdir(parents=True, exist_ok=True)
    WEB_STRUCT_ROOT.mkdir(parents=True, exist_ok=True)

    summary: Dict[str, int] = defaultdict(int)

    for dataset, list_key in DATASET_SPECS.items():
        rows = collect_dataset(dataset, list_key)
        csv_text = render_csv(rows)
        out_path = OUTPUT_ROOT / f"{dataset}.csv"
        out_path.write_text(csv_text, encoding="utf-8")
        web_path = WEB_STRUCT_ROOT / f"{dataset}.csv"
        web_path.write_text(csv_text, encoding="utf-8")
        summary[dataset] = len(rows)
        print(f"[ok] {dataset}: {len(rows)} rows -> {out_path}")

    missing = {k: v for k, v in summary.items() if v == 0}
    if missing:
        print("Datasets with no rows (likely no published data):")
        for name in sorted(missing):
            print(f" - {name}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
