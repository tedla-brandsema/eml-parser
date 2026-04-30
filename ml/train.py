from __future__ import annotations

import argparse
import csv
from dataclasses import dataclass
from pathlib import Path
import time

import torch
import torch.nn as nn
import torch.optim as optim

from ml.datasets.artifact_family_match import make_loader as make_artifact_loader
from ml.datasets.mock_family_match import FAMILIES, make_train_loader as make_mock_train_loader
from ml.models.baseline import Baseline
from ml.models.oracle import OracleFamilyModel
from ml.models.set_encoder import SampleSetEncoder
from ml.models.set_encoder_plus_stats import SampleSetEncoderPlusStats


DEVICE = torch.device("cuda" if torch.cuda.is_available() else "cpu")
print(f"Using device: {DEVICE}")

ML_DIR = Path(__file__).resolve().parent
CHECKPOINT_DIR = ML_DIR / "checkpoints"
RESULTS_DIR = ML_DIR / "results"
CHECKPOINT_DIR.mkdir(exist_ok=True)
RESULTS_DIR.mkdir(exist_ok=True)


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
    parser = argparse.ArgumentParser(description="Train family-match models for eml-parser")
    parser.add_argument(
        "--artifact-dir",
        type=Path,
        default=None,
        help="Directory of Go-generated family artifact JSON files under artifacts/equivalence/",
    )
    parser.add_argument("--epochs", type=int, default=8, help="Training epochs per model")
    parser.add_argument("--batch-size", type=int, default=64, help="Training batch size")
    parser.add_argument("--seed", type=int, default=0, help="Mock-data seed when artifact-dir is not used")
    return parser.parse_args()


def build_dataset_config(args: argparse.Namespace) -> DatasetConfig:
    if args.artifact_dir is None:
        loader = make_mock_train_loader(
            batch_size=args.batch_size,
            n_samples=4096,
            set_size=16,
            seed=args.seed,
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
    loader = make_artifact_loader(paths, batch_size=args.batch_size, shuffle=True)
    dataset = loader.dataset
    return DatasetConfig(
        name="artifact_family_match",
        loader=loader,
        set_size=dataset.set_size,
        d_point=dataset.d_point,
        n_classes=dataset.n_classes,
    )


def build_model(spec: ModelSpec, config: DatasetConfig) -> nn.Module:
    if spec.name == "baseline":
        return Baseline(set_size=config.set_size, d_point=config.d_point, n_classes=config.n_classes)
    if spec.name == "set_encoder":
        return SampleSetEncoder(d_point=config.d_point, hidden=64, n_classes=config.n_classes, agg="sum")
    if spec.name == "set_encoder_plus_stats":
        return SampleSetEncoderPlusStats(d_point=config.d_point, hidden=64, n_classes=config.n_classes, agg="sum")
    if spec.name == "oracle":
        return OracleFamilyModel(d_oracle=config.n_classes, hidden=32, n_classes=config.n_classes)
    raise ValueError(f"unsupported model spec: {spec.name}")


def train_model(
    model: nn.Module,
    loader,
    *,
    input_key: str,
    epochs: int,
    lr: float = 1e-3,
) -> float:
    model.to(DEVICE)
    model.train()
    opt = optim.Adam(model.parameters(), lr=lr)
    loss_fn = nn.CrossEntropyLoss()

    if DEVICE.type == "cuda":
        torch.cuda.synchronize()
    start = time.perf_counter()

    for epoch in range(epochs):
        last_loss = None
        for batch in loader:
            inputs = batch[input_key].to(DEVICE)
            y = batch["y"].to(DEVICE)

            opt.zero_grad(set_to_none=True)
            logits = model(inputs)
            loss = loss_fn(logits, y)
            loss.backward()
            opt.step()
            last_loss = loss

        if epoch == 0 or epoch == epochs - 1:
            value = float(last_loss.item()) if last_loss is not None else float("nan")
            print(f"  epoch {epoch+1}/{epochs}: loss={value:.4f}")

    if DEVICE.type == "cuda":
        torch.cuda.synchronize()
    end = time.perf_counter()
    return end - start


def main() -> None:
    args = parse_args()
    config = build_dataset_config(args)
    rows: list[dict[str, object]] = []

    print(
        f"Training dataset: {config.name} "
        f"(set_size={config.set_size}, d_point={config.d_point}, n_classes={config.n_classes})"
    )

    for spec in MODELS:
        print(f"\nTraining {spec.name} on {config.name}")
        model = build_model(spec, config)
        elapsed = train_model(model, config.loader, input_key=spec.input_key, epochs=args.epochs)
        ckpt = CHECKPOINT_DIR / f"{spec.name}__{config.name}.pt"
        torch.save(model.state_dict(), ckpt)
        print(f"Saved checkpoint: {ckpt}")
        print(f"Training time: {elapsed:.2f}s")

        rows.append(
            {
                "model": spec.name,
                "dataset": config.name,
                "train_time_sec": round(elapsed, 4),
                "device": str(DEVICE),
            }
        )

    out = RESULTS_DIR / "training_times.csv"
    with open(out, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=rows[0].keys())
        writer.writeheader()
        writer.writerows(rows)
    print(f"\nSaved training times to {out}")


if __name__ == "__main__":
    main()
