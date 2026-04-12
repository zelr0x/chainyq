import pandas as pd
import matplotlib.pyplot as plt
import numpy as np


df = pd.read_csv("bench.csv")

df["impl"] = df["impl"].str.strip().str.lower()
df["variant"] = df["variant"].str.strip().str.lower()

ALLOWED_BENCHES = {
    "PushBack",
    "PushFront",
    "RandomAccess",
    "Churn",
    "SlidingWindowMax",
    "BurstyQueue",
}

df = df[df["benchmark"].isin(ALLOWED_BENCHES)].copy()


df["key"] = df["impl"] + "_" + df["variant"]


order = [
    ("chainyq", "ensure"),
    ("chainyq", "base"),
    ("gammazero", "setbasecap"),
    ("gammazero", "base"),
    ("edwingeng", "base"),
]

order_keys = [f"{i}_{v}" for i, v in order]

labels = {
    "chainyq_ensure": "chainyq (ensure)",
    "chainyq_base": "chainyq",
    "gammazero_setbasecap": "gammazero (setbasecap)",
    "gammazero_base": "gammazero",
    "edwingeng_base": "edwingeng",
}

colors = {
    "chainyq_ensure": "#4c9be8",
    "chainyq_base": "#1f77b4",
    "gammazero_setbasecap": "#66c266",
    "gammazero_base": "#2ca02c",
    "edwingeng_base": "#ff7f0e",
}

benchmarks = df["benchmark"].unique()


pivot = df.pivot_table(
    index="benchmark",
    columns="key",
    values="ns_op",
    aggfunc="mean"
)
pivot = pivot.reindex(ALLOWED_BENCHES)


fig, ax = plt.subplots(figsize=(14, 7))

x = np.arange(len(benchmarks))
bar_width = 0.80

cap = 35

for bi, bench in enumerate(benchmarks):
    row = pivot.loc[bench]
    # keep only existing values for THIS benchmark
    available = [
        k for k in order_keys
        if k in row.index and not pd.isna(row[k])
    ]

    if not available:
        continue

    w = bar_width / len(available)

    for i, key in enumerate(available):
        xpos = bi - bar_width/2 + i*w
        ax.bar(
            xpos,
            row[key],
            width=w,
            color=colors.get(key, None),
            label=labels.get(key, key),
        )

        val = row[key]
        if val > cap:
            ax.text(
                xpos,
                cap * 0.95,
                val,
                ha="center",
                va="top",
                fontsize=10,
            )


ax.set_xticks(x)
ax.set_xticklabels(benchmarks)

ax.grid(True, which="both", axis="y", alpha=0.12, linestyle="--")

ax.set_ylim(0, cap)
ax.set_ylabel("ns/op")
ax.set_title(
    "Deque Benchmarks (lower is better)",
    fontsize=14,
    fontweight="medium",
    pad=12
)

# dedupe legend
handles, lbls = ax.get_legend_handles_labels()
unique = dict(zip(lbls, handles))
ax.legend(unique.values(), unique.keys())

ax.spines["top"].set_visible(False)
ax.spines["right"].set_visible(False)
ax.spines["left"].set_alpha(0.3)
ax.spines["bottom"].set_alpha(0.3)
ax.tick_params(axis="both", colors="#333333", labelsize=10)

plt.tight_layout()
plt.style.use("default")
plt.savefig("bench.png", dpi=300, bbox_inches="tight")

print("Saved bench.png")
