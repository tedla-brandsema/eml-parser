from __future__ import annotations

import argparse
import csv
from dataclasses import dataclass
from pathlib import Path
import random

import numpy as np
import torch

from ml.datasets.artifact_family_match import make_loader as make_artifact_loader
from ml.datasets.mock_family_match import FAMILIES, make_test_loader as make_mock_test_loader
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

EVAL_SEEDS = [0, 1, 2, 3, 4]


@dataclass(frozen=True)
class DatasetConfig:
    name: str
    loader: object
    set_size: int
    d_point: int
    n_classes: int


@dataclass(frozen=True)
class ModelSpec:
    name: str
    input_key: str


MODELS: tuple[ModelSpec, ...] = (
    ModelSpec("baseline", "x"),
    ModelSpec("set_encoder", "x"),
    ModelSpec("set_encoder_plus_stats", "x"),
    ModelSpec("oracle", "oracle"),
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Evaluate family-match models for eml-parser")
    parser.add_argument(
        "--artifact-dir",
        type=Path,
        default=None,
        help="Directory of Go-generated family artifact JSON files under artifacts/equivalence/",
    )
    parser.add_argument("--batch-size", type=int, default=256, help="Evaluation batch size")
    return parser.parse_args()


def build_dataset_config(args: argparse.Namespace, *, seed: int) -> DatasetConfig:
    if args.artifact_dir is None:
        loader = make_mock_test_loader(
            batch_size=args.batch_size,
            n_samples=1024,
            set_size=16,
            seed=seed,
        )
        return DatasetConfig(
            name="mock_family_match",
            loader=loader,
            set_size=16,
            d_point=2,
            n_classes=len(FAMILIES),
        )

    artifact_dir = args.artifact_dir.resolve()
    paths = sorted(p for p in artifact_dir.glob("*.json") if not p.name.endswith(".family.json"))
    if not paths:
        raise ValueError(f"no artifact JSON files found under {artifact_dir}")
    loader = make_artifact_loader(paths, batch_size=args.batch_size, shuffle=False)
    dataset = loader.dataset
    return DatasetConfig(
        name="artifact_family_match",
        loader=loader,
        set_size=dataset.set_size,
        d_point=dataset.d_point,
        n_classes=dataset.n_classes,
    )


def build_model(spec: ModelSpec, config: DatasetConfig) -> torch.nn.Module:
    if spec.name == "baseline":
        return Baseline(set_size=config.set_size, d_point=config.d_point, n_classes=config.n_classes)
    if spec.name == "set_encoder":
        return SampleSetEncoder(d_point=config.d_point, hidden=64, n_classes=config.n_classes, agg="sum")
    if spec.name == "set_encoder_plus_stats":
        return SampleSetEncoderPlusStats(d_point=config.d_point, hidden=64, n_classes=config.n_classes, agg="sum")
    if spec.name == "oracle":
        return OracleFamilyModel(d_oracle=config.n_classes, hidden=32, n_classes=config.n_classes)
    raise ValueError(f"unsupported model spec: {spec.name}")


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
    args = parse_args()
    train_times = load_training_times()
    rows: list[dict[str, object]] = []

    for seed in EVAL_SEEDS:
        set_seed(seed)
        config = build_dataset_config(args, seed=seed)
        if seed == EVAL_SEEDS[0]:
            print(
                f"Evaluation dataset: {config.name} "
                f"(set_size={config.set_size}, d_point={config.d_point}, n_classes={config.n_classes})"
            )

        seed_rows: list[dict[str, object]] = []
        oracle_acc = None

        for spec in MODELS:
            model = build_model(spec, config).to(DEVICE)
            load_checkpoint(model, spec.name, config.name)

            row: dict[str, object] = {
                "model": spec.name,
                "dataset": config.name,
                "seed": seed,
                "family_accuracy": family_accuracy(model, config.loader, input_key=spec.input_key),
                "top3_family_accuracy": topk_family_accuracy(model, config.loader, input_key=spec.input_key, k=3),
                "train_time_sec": train_times.get((spec.name, config.name)),
            }

            perm = permutation_sensitivity(model, config.loader, input_key=spec.input_key)
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
