from __future__ import annotations

import torch
import torch.nn as nn

from ml.models.set_encoder import SampleSetEncoder


class SampleSetEncoderPlusStats(nn.Module):
    """
    Set encoder augmented with simple summary statistics.
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
        self.encoder = SampleSetEncoder(
            d_point=d_point,
            hidden=hidden,
            n_classes=hidden,
            agg=agg,
        )
        self.head = nn.Sequential(
            nn.Linear(hidden + 3 * d_point, hidden),
            nn.ReLU(),
            nn.Linear(hidden, n_classes),
        )

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        q = self.encoder(x)
        mean = x.mean(dim=1)
        var = x.var(dim=1, unbiased=False)
        summ = x.sum(dim=1)
        feats = torch.cat([q, summ, mean, var], dim=-1)
        return self.head(feats)
