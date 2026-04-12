import re
import csv
import sys


input_file = sys.argv[1] if len(sys.argv) > 1 else "../bench.txt"
output_file = "bench.csv"


LINE_RE = re.compile(
    r"^Benchmark(?P<op>[^/]+)/(?P<impl>.+?)-\d+"
    r".*?(?P<ns>[0-9.]+)\s+ns/op"
    r".*?(?P<b>[0-9.]+)\s+B/op"
    r".*?(?P<alloc>\d+)\s+allocs/op"
)


def normalize_impl(raw: str):
    if "chainyq" in raw:
        if ".list" in raw or "/list" in raw:
            return None, None
        base = "chainyq"
    elif "edwingeng" in raw:
        base = "edwingeng"
    elif "gammazero" in raw:
        base = "gammazero"
    else:
        return None, None

    if "Pooled" in raw or "Reserve" in raw:
        return None, None

    variant = "base"

    if "Ensure" in raw:
        variant = "ensure"
    elif "SetBaseCap" in raw:
        variant = "setbasecap"

    return base, variant


rows = []

with open(input_file, "r", encoding="utf-8") as f:
    for line in f:
        m = LINE_RE.match(line)
        if not m:
            continue

        op = m.group("op")
        impl_raw = m.group("impl")

        base, variant = normalize_impl(impl_raw)
        if base is None:
            continue

        rows.append([
            op,
            base,
            variant,
            float(m.group("ns")),
            float(m.group("b")),
            int(m.group("alloc")),
        ])


with open(output_file, "w", newline="") as f:
    w = csv.writer(f)
    w.writerow(["benchmark", "impl", "variant", "ns_op", "b_op", "allocs_op"])
    w.writerows(rows)

print(f"Wrote {output_file} ({len(rows)} rows)")
