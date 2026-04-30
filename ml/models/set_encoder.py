from __future__ import annotations

import torch
import torch.nn as nn


class SampleSetEncoder(nn.Module):
    """
    DeepSets-style encoder over sampled `(x, y)` points.
    """

    def __init__(
        self,
        *,
        d_point: int,
        hidden: int = 64,
        n_classes: int = 4,
        agg: str = "sum",
    ) -> None:
        super().__init__()
        if agg not in {"sum", "mean", "max"}:
            raise ValueError(f"unsupported agg: {agg}")
        self.agg = agg

        self.psi = nn.Sequential(
            nn.Linear(d_point, hidden),
            nn.ReLU(),
            nn.Linear(hidden, hidden),
            nn.ReLU(),
        )
        self.rho = nn.Linear(hidden, n_classes)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        h = self.psi(x)
        if self.agg == "sum":
            pooled = h.sum(dim=1)
        elif self.agg == "mean":
            pooled = h.mean(dim=1)
        else:
            pooled = h.max(dim=1).values
        return self.rho(pooled)
