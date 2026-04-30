from __future__ import annotations

import json
from pathlib import Path

import torch

from ml.datasets.artifact_family_match import ArtifactFamilyMatchDataset, make_loader


def write_artifact(path: Path, *, family_id: int, n_families: int, points: list[list[float]]) -> None:
    payload = {
        "family_id": family_id,
        "family_name": f"family_{family_id}",
        "n_families": n_families,
        "samples": [
            {
                "points": points,
                "family_id": family_id,
                "oracle": [1.0 if i == family_id else 0.0 for i in range(n_families)],
            }
        ],
    }
    path.write_text(json.dumps(payload))


def test_artifact_dataset_reads_metadata(tmp_path: Path) -> None:
    write_artifact(
        tmp_path / "a.json",
        family_id=0,
        n_families=2,
        points=[[0.0, 0.0], [1.0, 1.0]],
    )
    write_artifact(
        tmp_path / "b.json",
        family_id=1,
        n_families=2,
        points=[[0.5, 0.25], [1.5, 2.25]],
    )

    ds = ArtifactFamilyMatchDataset(tmp_path.glob("*.json"))
    assert len(ds) == 2
    assert ds.n_classes == 2
    assert ds.set_size == 2
    assert ds.d_point == 2

    item = ds[0]
    assert item["x"].shape == (2, 2)
    assert item["y"].dtype == torch.long
    assert item["oracle"].shape == (2,)


def test_artifact_loader_batches(tmp_path: Path) -> None:
    write_artifact(
        tmp_path / "family.json",
        family_id=0,
        n_families=1,
        points=[[0.0, 1.0], [1.0, 2.0], [2.0, 3.0]],
    )
    loader = make_loader(tmp_path.glob("*.json"), batch_size=1, shuffle=False)
    batch = next(iter(loader))
    assert batch["x"].shape == (1, 3, 2)
    assert batch["y"].shape == (1,)
    assert batch["oracle"].shape == (1, 1)
