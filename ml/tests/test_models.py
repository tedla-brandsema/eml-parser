from __future__ import annotations

import torch

from ml.datasets.mock_family_match import FAMILIES
from ml.models.baseline import Baseline
from ml.models.oracle import OracleFamilyModel
from ml.models.set_encoder import SampleSetEncoder
from ml.models.set_encoder_plus_stats import SampleSetEncoderPlusStats


def test_model_output_shapes():
    batch = torch.randn(4, 16, 2)
    oracle = torch.randn(4, len(FAMILIES))

    baseline = Baseline(set_size=16, d_point=2, n_classes=len(FAMILIES))
    set_encoder = SampleSetEncoder(d_point=2, hidden=32, n_classes=len(FAMILIES))
    plus_stats = SampleSetEncoderPlusStats(d_point=2, hidden=32, n_classes=len(FAMILIES))
    oracle_model = OracleFamilyModel(d_oracle=len(FAMILIES), hidden=16, n_classes=len(FAMILIES))

    assert baseline(batch).shape == (4, len(FAMILIES))
    assert set_encoder(batch).shape == (4, len(FAMILIES))
    assert plus_stats(batch).shape == (4, len(FAMILIES))
    assert oracle_model(oracle).shape == (4, len(FAMILIES))
