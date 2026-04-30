from __future__ import annotations

import torch
import torch.nn as nn


class OracleFamilyModel(nn.Module):
    """
    Upper-bound model that receives true family-level oracle features directly.
    """

    def __init__(self, *, d_oracle: int, hidden: int = 32, n_classes: int = 4) -> None:
        super().__init__()
        self.net = nn.Sequential(
            nn.Linear(d_oracle, hidden),
            nn.ReLU(),
            nn.Linear(hidden, n_classes),
        )

    def forward(self, oracle: torch.Tensor) -> torch.Tensor:
        return self.net(oracle)
