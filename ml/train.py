from __future__ import annotations

import csv
from dataclasses import dataclass
from pathlib import Path
import time

import torch
import torch.nn as nn
import torch.optim as optim

from ml.datasets.mock_family_match import FAMILIES, make_train_loader
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

SET_SIZE = 16
D_POINT = 2
N_CLASSES = len(FAMILIES)


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


def train_model(
    model: nn.Module,
    loader,
    *,
    input_key: str,
    epochs: int = 8,
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
    rows: list[dict[str, object]] = []
    loader = make_train_loader(batch_size=64, n_samples=4096, set_size=SET_SIZE, seed=0)

    for spec in MODELS:
        print(f"\nTraining {spec.name} on mock_family_match")
        model = spec.factory()
        elapsed = train_model(model, loader, input_key=spec.input_key)
        ckpt = CHECKPOINT_DIR / f"{spec.name}__mock_family_match.pt"
        torch.save(model.state_dict(), ckpt)
        print(f"Saved checkpoint: {ckpt}")
        print(f"Training time: {elapsed:.2f}s")

        rows.append(
            {
                "model": spec.name,
                "dataset": "mock_family_match",
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
