from __future__ import annotations

import torch
from torch.utils.data import DataLoader, Dataset

from ml.metrics import family_accuracy, permutation_sensitivity, topk_family_accuracy
from ml.models.baseline import Baseline
from ml.models.set_encoder import SampleSetEncoder


class TinyDataset(Dataset):
    def __init__(self) -> None:
        self.items = [
            {
                "x": torch.tensor([[0.0, 0.0], [1.0, 1.0]], dtype=torch.float32),
                "y": torch.tensor(0, dtype=torch.long),
                "oracle": torch.tensor([1.0, 0.0], dtype=torch.float32),
            },
            {
                "x": torch.tensor([[0.0, 1.0], [1.0, 2.0]], dtype=torch.float32),
                "y": torch.tensor(1, dtype=torch.long),
                "oracle": torch.tensor([0.0, 1.0], dtype=torch.float32),
            },
        ]

    def __len__(self) -> int:
        return len(self.items)

    def __getitem__(self, idx: int):
        return self.items[idx]


def test_metrics_run():
    loader = DataLoader(TinyDataset(), batch_size=2)
    baseline = Baseline(set_size=2, d_point=2, n_classes=2)
    set_model = SampleSetEncoder(d_point=2, hidden=8, n_classes=2)

    acc = family_accuracy(baseline, loader, input_key="x")
    top3 = topk_family_accuracy(baseline, loader, input_key="x", k=3)
    perm = permutation_sensitivity(set_model, loader, input_key="x", n_batches=1, n_perms=2)

    assert 0.0 <= acc <= 1.0
    assert 0.0 <= top3 <= 1.0
    assert perm is not None
    assert perm >= 0.0
