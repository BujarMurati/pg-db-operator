import pytest


@pytest.fixture
def env(monkeypatch) -> None:
    """Set environment variables for connecting to the database"""
    monkeypatch.setenv("PGHOST", "localhost")
    monkeypatch.setenv("PGPASSWORD", "test")
    monkeypatch.setenv("PGUSER", "server_admin")
    monkeypatch.setenv("PGPORT", "5432")
    monkeypatch.setenv("PGDATABASE", "postgres")
