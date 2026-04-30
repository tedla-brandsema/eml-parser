from __future__ import annotations

import torch
import torch.nn as nn


class Baseline(nn.Module):
    """
    Order-sensitive baseline over flattened sampled points.
    """

    def __init__(self, *, set_size: int, d_point: int, n_classes: int, hidden: int = 64) -> None:
        super().__init__()
        self.net = nn.Sequential(
            nn.Flatten(),
            nn.Linear(set_size * d_point, hidden),
            nn.ReLU(),
            nn.Linear(hidden, n_classes),
        )

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        return self.net(x)
