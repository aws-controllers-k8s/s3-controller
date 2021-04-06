import pytest
from e2e import SERVICE_NAME

class TestBucket:
    def test_bucket(self):
        pytest.skip(f"No tests for {SERVICE_NAME}")