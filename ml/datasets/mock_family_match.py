from __future__ import annotations

from dataclasses import dataclass
import math
import random
from typing import Callable

import torch
from torch.utils.data import DataLoader, Dataset


FamilyFn = Callable[[torch.Tensor], torch.Tensor]


@dataclass(frozen=True)
class FamilySpec:
    family_id: int
    name: str
    members: tuple[str, ...]
    fn: FamilyFn


def _sin_pi(x: torch.Tensor) -> torch.Tensor:
    return torch.sin(math.pi * x)


FAMILIES: tuple[FamilySpec, ...] = (
    FamilySpec(0, "linear", ("x", "add_zero", "sub_neg_zero"), lambda x: x),
    FamilySpec(1, "quadratic", ("square", "mul_self", "pow_two"), lambda x: x * x),
    FamilySpec(2, "sin_pi", ("sin_pi", "sin_shift_wrap", "periodic_alt"), _sin_pi),
    FamilySpec(
        3,
        "reciprocal_shift",
        ("recip_shift", "div_one_shift", "pow_neg_one_shift"),
        lambda x: 1.0 / (1.5 + x),
    ),
)


class MockFamilyMatchDataset(Dataset):
    """
    Mock dataset for the first equivalence-family task.

    Each sample is a set of observed `(x, y)` points drawn from one hidden law.
    The target is the family id, not the exact member variant.
    """

    def __init__(
        self,
        *,
        n_samples: int = 4096,
        set_size: int = 16,
        x_min: float = -0.75,
        x_max: float = 0.75,
        seed: int = 0,
        split: str = "train",
    ) -> None:
        super().__init__()
        if split not in {"train", "test"}:
            raise ValueError(f"unsupported split: {split}")

        rng = random.Random(seed + (0 if split == "train" else 10_000))
        torch_gen = torch.Generator().manual_seed(seed + (0 if split == "train" else 10_000))
        self.samples: list[dict[str, torch.Tensor | int | str]] = []

        for _ in range(n_samples):
            family = rng.choice(FAMILIES)
            member_name = rng.choice(family.members)

            xs = torch.empty(set_size, dtype=torch.float32).uniform_(x_min, x_max, generator=torch_gen)
            ys = family.fn(xs)
            points = torch.stack([xs, ys], dim=-1)

            perm = torch.randperm(set_size, generator=torch_gen)
            points = points[perm]

            oracle = torch.nn.functional.one_hot(
                torch.tensor(family.family_id, dtype=torch.long),
                num_classes=len(FAMILIES),
            ).to(torch.float32)

            self.samples.append(
                {
                    "x": points,
                    "y": family.family_id,
                    "oracle": oracle,
                    "member_name": member_name,
                    "family_name": family.name,
                }
            )

    def __len__(self) -> int:
        return len(self.samples)

    def __getitem__(self, idx: int) -> dict[str, torch.Tensor]:
        item = self.samples[idx]
        return {
            "x": item["x"],
            "y": torch.tensor(item["y"], dtype=torch.long),
            "oracle": item["oracle"],
        }


def make_train_loader(
    *,
    batch_size: int = 64,
    n_samples: int = 4096,
    set_size: int = 16,
    seed: int = 0,
) -> DataLoader:
    ds = MockFamilyMatchDataset(
        n_samples=n_samples,
        set_size=set_size,
        seed=seed,
        split="train",
    )
    return DataLoader(ds, batch_size=batch_size, shuffle=True)


def make_test_loader(
    *,
    batch_size: int = 256,
    n_samples: int = 1024,
    set_size: int = 16,
    seed: int = 0,
) -> DataLoader:
    ds = MockFamilyMatchDataset(
        n_samples=n_samples,
        set_size=set_size,
        seed=seed,
        split="test",
    )
    return DataLoader(ds, batch_size=batch_size, shuffle=False)
