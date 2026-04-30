from __future__ import annotations

import json
from pathlib import Path
from typing import Iterable

import torch
from torch.utils.data import DataLoader, Dataset


class ArtifactFamilyMatchDataset(Dataset):
    """
    Reader for future Go-generated family-match corpora.

    Item 34 keeps this intentionally narrow. The loader expects each JSON file
    to already contain sampled point sets plus a family label and optional oracle
    features. The exact long-term schema is deferred to item 36.
    """

    def __init__(self, paths: Iterable[Path]) -> None:
        self.samples: list[dict[str, torch.Tensor]] = []
        for path in paths:
            with open(path) as f:
                payload = json.load(f)

            if "samples" not in payload:
                raise ValueError(f"{path} is missing 'samples'")

            for sample in payload["samples"]:
                x = torch.tensor(sample["points"], dtype=torch.float32)
                y = torch.tensor(sample["family_id"], dtype=torch.long)
                oracle = sample.get("oracle")
                if oracle is None:
                    oracle_tensor = torch.nn.functional.one_hot(
                        y,
                        num_classes=payload.get("n_families", int(y.item()) + 1),
                    ).to(torch.float32)
                else:
                    oracle_tensor = torch.tensor(oracle, dtype=torch.float32)

                self.samples.append({"x": x, "y": y, "oracle": oracle_tensor})

    def __len__(self) -> int:
        return len(self.samples)

    def __getitem__(self, idx: int) -> dict[str, torch.Tensor]:
        return self.samples[idx]


def make_loader(paths: Iterable[Path], *, batch_size: int = 256, shuffle: bool = False) -> DataLoader:
    ds = ArtifactFamilyMatchDataset(paths)
    return DataLoader(ds, batch_size=batch_size, shuffle=shuffle)
