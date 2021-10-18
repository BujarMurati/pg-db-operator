from pg_db_operator.app import create_app
from starlette.testclient import TestClient


def test_health_check_passes_with_working_database_connection(env):
    client = TestClient(create_app())
    response = client.get("/health")
    assert response.ok


def test_health_check_fails_with_failing_database_connection():
    client = TestClient(create_app())
    response = client.get("/health")
    assert not response.ok
    assert 400 <= response.status_code < 600
