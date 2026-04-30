from __future__ import annotations

from typing import Iterable

import numpy as np
import torch


def _to_model_device(tensor: torch.Tensor, model: torch.nn.Module) -> torch.Tensor:
    return tensor.to(next(model.parameters()).device)


def _predict(model: torch.nn.Module, batch: dict[str, torch.Tensor], input_key: str) -> torch.Tensor:
    inputs = _to_model_device(batch[input_key], model)
    return model(inputs)


@torch.no_grad()
def family_accuracy(model: torch.nn.Module, loader: Iterable[dict[str, torch.Tensor]], *, input_key: str) -> float:
    model.eval()
    correct = 0
    total = 0
    for batch in loader:
        logits = _predict(model, batch, input_key)
        y = _to_model_device(batch["y"], model)
        pred = logits.argmax(dim=-1)
        correct += (pred == y).sum().item()
        total += y.numel()
    return correct / total if total else 0.0


@torch.no_grad()
def topk_family_accuracy(
    model: torch.nn.Module,
    loader: Iterable[dict[str, torch.Tensor]],
    *,
    input_key: str,
    k: int = 3,
) -> float:
    model.eval()
    correct = 0
    total = 0
    for batch in loader:
        logits = _predict(model, batch, input_key)
        y = _to_model_device(batch["y"], model)
        topk = logits.topk(k=min(k, logits.shape[-1]), dim=-1).indices
        correct += (topk == y.unsqueeze(-1)).any(dim=-1).sum().item()
        total += y.numel()
    return correct / total if total else 0.0


@torch.no_grad()
def permutation_sensitivity(
    model: torch.nn.Module,
    loader: Iterable[dict[str, torch.Tensor]],
    *,
    input_key: str,
    n_batches: int = 5,
    n_perms: int = 5,
) -> float | None:
    if input_key != "x":
        return None

    model.eval()
    diffs: list[float] = []
    for bi, batch in enumerate(loader):
        if bi >= n_batches:
            break
        x = _to_model_device(batch["x"], model)
        base = model(x)
        _, n_points, _ = x.shape
        for _ in range(n_perms):
            perm = torch.randperm(n_points, device=x.device)
            out = model(x[:, perm, :])
            diffs.append((out - base).abs().mean().item())
    return float(np.mean(diffs)) if diffs else 0.0
