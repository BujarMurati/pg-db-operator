import time
import json
from base64 import b64decode
import subprocess
import pytest
from kopf.testing import KopfRunner
from pg_db_operator.database import DatabaseServer

db = DatabaseServer()

# https://kopf.readthedocs.io/en/stable/troubleshooting/#kubectl-freezes-on-object-deletion
REMOVE_KOPF_FINALIZER = """
kubectl patch pgdb pg-db-test -p '{"metadata": {"finalizers": []}}' \
    --type merge -n testing
"""


@pytest.fixture
def setup():
    try:
        subprocess.run("kubectl create ns testing", shell=True, check=True)
        yield
    finally:
        subprocess.run(
            REMOVE_KOPF_FINALIZER,
            shell=True,
            check=False,
        )
        subprocess.run("kubectl delete ns testing", shell=True, check=True)


def _decode_secret_data(value: str) -> str:
    return b64decode(value.encode()).decode()


def get_secret(name: str = "secret-pg-db-test") -> str:
    return subprocess.run(
        f"kubectl get secret {name} -n testing -o json",
        shell=True,
        check=True,
        text=True,
        capture_output=True,
    ).stdout


@pytest.mark.asyncio
async def test_operator_creates_database(setup, env):
    with KopfRunner(["run", "--verbose", "pg_db_operator/operator.py"]) as runner:
        assert not await db.database_exists("test_db")
        assert not await db.user_exists("test_db")
        assert not await db.user_has_all_privileges_on_database("test_db")
        subprocess.run("kubectl apply -f tests/data/pg-db.yaml -n testing", shell=True, check=True)
        time.sleep(2)
        assert await db.database_exists("test_db")
        assert await db.user_exists("test_db")
        assert await db.user_has_all_privileges_on_database("test_db")

    assert runner.exit_code == 0
    assert runner.exception is None


@pytest.mark.asyncio
async def test_operator_creates_secret(setup, env):
    with KopfRunner(["run", "--verbose", "pg_db_operator/operator.py"]) as runner:
        subprocess.run("kubectl apply -f tests/data/pg-db.yaml -n testing", shell=True, check=True)
        time.sleep(2)
        secret = get_secret()
        secret_data = json.loads(secret)["data"]
        assert "PGDATABASE" in secret_data
        assert "PGPASSWORD" in secret_data
        assert "PGUSER" in secret_data
        assert _decode_secret_data(secret_data["PGUSER"]) == "test_db"
        assert _decode_secret_data(secret_data["PGDATABASE"]) == "test_db"
    assert runner.exit_code == 0
    assert runner.exception is None


@pytest.mark.asyncio
async def test_deletes_secret_but_keeps_database(setup, env):
    with KopfRunner(["run", "--verbose", "pg_db_operator/operator.py"]) as runner:
        subprocess.run("kubectl apply -f tests/data/pg-db.yaml -n testing", shell=True, check=True)
        subprocess.run("kubectl delete -f tests/data/pg-db.yaml -n testing", shell=True, check=True)
        time.sleep(1)
        with pytest.raises(subprocess.CalledProcessError):
            get_secret()
        assert await db.database_exists("test_db")
        assert await db.user_exists("test_db")
        assert await db.user_has_all_privileges_on_database("test_db")
    assert runner.exit_code == 0
    assert runner.exception is None


@pytest.mark.asyncio
async def test_update_recreates_secret_database(setup, env):
    with KopfRunner(["run", "--verbose", "pg_db_operator/operator.py"]) as runner:
        subprocess.run("kubectl apply -f tests/data/pg-db.yaml -n testing", shell=True, check=True)
        subprocess.run(
            """
            kubectl patch pgdb pg-db-test -p '{"spec": {"targetSecret": {"userNamePostfix": "@localhost"}}}' \
                -n testing --type merge
            """,
            shell=True,
            check=True,
        )
        time.sleep(1)

        secret = get_secret()
        print(secret)
        secret_data = json.loads(secret)["data"]
        assert _decode_secret_data(secret_data["PGUSER"]) == "test_db@localhost"
    assert runner.exit_code == 0
    assert runner.exception is None
