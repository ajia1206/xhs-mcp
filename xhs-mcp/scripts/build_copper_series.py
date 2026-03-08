#!/usr/bin/env python3
"""
Build a daily copper futures closing price series for October 2025.

The script reads the normalized CSV (`structured/futures_kx.csv`), selects the
contract with the highest open interest per trading day (excluding subtotal
rows), and emits a compact JSON file consumable by the visualization page.
"""

from __future__ import annotations

import csv
import json
from collections import defaultdict
from pathlib import Path
from typing import Dict, List

INPUT_CSV = Path("data/shfe/2025-10/structured/futures_kx.csv")
OUTPUT_JSON = Path("web/data/cu_main_series.json")


def parse_csv(path: Path) -> List[Dict]:
    """Return rows for copper futures keyed by trade_date with numeric fields."""
    grouped: Dict[str, Dict] = defaultdict(dict)
    with path.open(encoding="utf-8") as fh:
        reader = csv.DictReader(fh)
        for row in reader:
            if row.get("PRODUCTID") != "cu_f":
                continue
            delivery = row.get("DELIVERYMONTH", "").strip()
            if not delivery or delivery == "小计":
                continue
            trade_date = row["trade_date"]
            open_interest = float(row["OPENINTEREST"]) if row["OPENINTEREST"] else 0.0
            existing = grouped[trade_date].get("row")
            close_price = float(row["CLOSEPRICE"]) if row["CLOSEPRICE"] else None
            if existing is None or open_interest > existing["open_interest"]:
                grouped[trade_date]["row"] = {
                    "trade_date": trade_date,
                    "delivery_month": delivery,
                    "close_price": close_price,
                    "open_interest": open_interest,
                    "volume": float(row["VOLUME"]) if row["VOLUME"] else None,
                }
    # Convert to sorted list
    series = [
        grouped[key]["row"]
        for key in sorted(grouped.keys())
        if grouped[key].get("row")
    ]
    # Normalize date string to ISO format for easier charting
    filtered = []
    for item in series:
        if item["close_price"] is None:
            continue
        d = item["trade_date"]
        item["date_iso"] = f"{d[:4]}-{d[4:6]}-{d[6:]}"
        filtered.append(item)
    return filtered


def main() -> int:
    if not INPUT_CSV.exists():
        raise SystemExit(f"Input CSV not found: {INPUT_CSV}")
    series = parse_csv(INPUT_CSV)
    OUTPUT_JSON.parent.mkdir(parents=True, exist_ok=True)
    OUTPUT_JSON.write_text(json.dumps(series, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"[ok] wrote {len(series)} points -> {OUTPUT_JSON}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
