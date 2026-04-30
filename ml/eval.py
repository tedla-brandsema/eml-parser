from __future__ import annotations

import csv
from dataclasses import dataclass
from pathlib import Path
import random

import numpy as np
import torch

from ml.datasets.mock_family_match import FAMILIES, make_test_loader
from ml.metrics import family_accuracy, permutation_sensitivity, topk_family_accuracy
from ml.models.baseline import Baseline
from ml.models.oracle import OracleFamilyModel
from ml.models.set_encoder import SampleSetEncoder
from ml.models.set_encoder_plus_stats import SampleSetEncoderPlusStats


DEVICE = torch.device("cuda" if torch.cuda.is_available() else "cpu")
print(f"Using device: {DEVICE}")

ML_DIR = Path(__file__).resolve().parent
CHECKPOINT_DIR = ML_DIR / "checkpoints"
RESULTS_DIR = ML_DIR / "results"
RESULTS_DIR.mkdir(exist_ok=True)

SET_SIZE = 16
D_POINT = 2
N_CLASSES = len(FAMILIES)
EVAL_SEEDS = [0, 1, 2, 3, 4]


@dataclass(frozen=True)
class ModelSpec:
    name: str
    input_key: str
    factory: callable


MODELS: tuple[ModelSpec, ...] = (
    ModelSpec(
        "baseline",
        "x",
        lambda: Baseline(set_size=SET_SIZE, d_point=D_POINT, n_classes=N_CLASSES),
    ),
    ModelSpec(
        "set_encoder",
        "x",
        lambda: SampleSetEncoder(d_point=D_POINT, hidden=64, n_classes=N_CLASSES, agg="sum"),
    ),
    ModelSpec(
        "set_encoder_plus_stats",
        "x",
        lambda: SampleSetEncoderPlusStats(d_point=D_POINT, hidden=64, n_classes=N_CLASSES, agg="sum"),
    ),
    ModelSpec(
        "oracle",
        "oracle",
        lambda: OracleFamilyModel(d_oracle=N_CLASSES, hidden=32, n_classes=N_CLASSES),
    ),
)


def set_seed(seed: int) -> None:
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)
    if torch.cuda.is_available():
        torch.cuda.manual_seed_all(seed)


def load_checkpoint(model: torch.nn.Module, model_name: str, dataset_name: str) -> bool:
    ckpt = CHECKPOINT_DIR / f"{model_name}__{dataset_name}.pt"
    if not ckpt.exists():
        print(f"[WARN] No checkpoint for {model_name} on {dataset_name}")
        return False
    model.load_state_dict(torch.load(ckpt, map_location=DEVICE, weights_only=True))
    return True


def load_training_times() -> dict[tuple[str, str], float]:
    path = RESULTS_DIR / "training_times.csv"
    if not path.exists():
        return {}

    out: dict[tuple[str, str], float] = {}
    with open(path) as f:
        for row in csv.DictReader(f):
            out[(row["model"], row["dataset"])] = float(row["train_time_sec"])
    return out


def main() -> None:
    train_times = load_training_times()
    rows: list[dict[str, object]] = []

    for seed in EVAL_SEEDS:
        set_seed(seed)
        loader = make_test_loader(batch_size=256, n_samples=1024, set_size=SET_SIZE, seed=seed)

        seed_rows: list[dict[str, object]] = []
        oracle_acc = None

        for spec in MODELS:
            model = spec.factory().to(DEVICE)
            load_checkpoint(model, spec.name, "mock_family_match")

            row: dict[str, object] = {
                "model": spec.name,
                "dataset": "mock_family_match",
                "seed": seed,
                "family_accuracy": family_accuracy(model, loader, input_key=spec.input_key),
                "top3_family_accuracy": topk_family_accuracy(model, loader, input_key=spec.input_key, k=3),
                "train_time_sec": train_times.get((spec.name, "mock_family_match")),
            }

            perm = permutation_sensitivity(model, loader, input_key=spec.input_key)
            if perm is not None:
                row["permutation_sensitivity"] = perm

            if spec.name == "oracle":
                oracle_acc = float(row["family_accuracy"])

            seed_rows.append(row)

        if oracle_acc is None:
            raise RuntimeError("oracle model metrics missing")

        for row in seed_rows:
            row["oracle_gap"] = oracle_acc - float(row["family_accuracy"])
            rows.append(row)

    fieldnames = sorted({k for row in rows for k in row.keys()})
    out = RESULTS_DIR / "results.csv"
    with open(out, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows)

    print(f"\nSaved results to {out}")
    for row in rows:
        print(row)


if __name__ == "__main__":
    main()
