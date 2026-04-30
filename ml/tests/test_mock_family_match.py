from __future__ import annotations

from ml.datasets.mock_family_match import FAMILIES, MockFamilyMatchDataset


def test_mock_family_match_shapes_and_labels():
    ds = MockFamilyMatchDataset(n_samples=8, set_size=6, seed=0, split="train")
    item = ds[0]

    assert item["x"].shape == (6, 2)
    assert item["y"].ndim == 0
    assert item["oracle"].shape == (len(FAMILIES),)
    assert 0 <= int(item["y"].item()) < len(FAMILIES)
