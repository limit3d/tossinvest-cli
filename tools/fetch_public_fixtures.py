#!/usr/bin/env python3
"""Fetch public Toss Securities web fixtures for local development."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from urllib.request import Request, urlopen


def specs(symbol: str) -> list[tuple[str, str]]:
    return [
        ("guest-init.json", "https://wts-api.tossinvest.com/api/v3/init?tabId=tab-bootstrap"),
        (
            "dashboard-trading-info.json",
            "https://wts-info-api.tossinvest.com/api/v1/dashboard/wts/overview/trading-info",
        ),
        ("stock-info.json", f"https://wts-info-api.tossinvest.com/api/v2/stock-infos/{symbol}"),
        (
            "stock-detail-common.json",
            f"https://wts-info-api.tossinvest.com/api/v1/stock-detail/ui/{symbol}/common",
        ),
        (
            "stock-price.json",
            f"https://wts-info-api.tossinvest.com/api/v1/product/stock-prices?meta=true&productCodes={symbol}",
        ),
        (
            "chart-day-1.json",
            f"https://wts-info-api.tossinvest.com/api/v1/c-chart/kr-s/{symbol}/day:1"
            "?count=61&session=all&investMode=integrated&useAdjustedRate=true",
        ),
    ]


def fetch(url: str) -> dict:
    request = Request(url, headers={"User-Agent": "tossctl-fixture-bootstrap/0.1"})
    with urlopen(request, timeout=20) as response:
        charset = response.headers.get_content_charset() or "utf-8"
        payload = response.read().decode(charset)
    return json.loads(payload)


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("output_dir", type=Path)
    parser.add_argument("--symbol", default="A005930")
    args = parser.parse_args(argv[1:])

    output_dir = args.output_dir
    output_dir.mkdir(parents=True, exist_ok=True)

    manifest = []
    for name, url in specs(args.symbol):
        data = fetch(url)
        path = output_dir / name
        with path.open("w", encoding="utf-8") as fp:
            json.dump(data, fp, ensure_ascii=False, indent=2)
            fp.write("\n")
        manifest.append({"file": name, "url": url})

    with (output_dir / "manifest.json").open("w", encoding="utf-8") as fp:
        json.dump(
            {
                "symbol": args.symbol,
                "fixtures": manifest,
            },
            fp,
            ensure_ascii=False,
            indent=2,
        )
        fp.write("\n")

    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
