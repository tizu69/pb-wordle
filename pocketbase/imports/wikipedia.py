#!/usr/bin/env python
# This script takes wikipedia.gz and turns it into a json.
# Grab a dump from https://dumps.wikimedia.org/other/pagetitles

import gzip, json, random

with gzip.open("wikipedia.gz", "rt") as f:
    rawlines = f.readlines()
    lines = [
        {"word": line.strip().replace("_", " ")}
        for line in rawlines
        if len(line.strip()) > 0 and line.strip().isascii() and random.random() < 0.003
    ]
    lines.pop(0)
    print("done processing (stripped some), out:", len(lines), "/", len(rawlines) - 1)
    json.dump(
        {"name": "Wikipedle", "mustHint": False, "mustPresent": False, "words": lines},
        open("wikipedia.json", "w"),
    )
