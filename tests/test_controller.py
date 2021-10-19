from base64 import b64encode
from datetime import datetime, timedelta
from typing import Optional
import pytest
from unittest.mock import AsyncMock
from pg_db_operator.controller import Controller, Spec, ObservedState, Secret, Status
from pg_db_operator.database import DatabaseServer


def mock_db(db: bool = False, user: bool = False, priv: bool = False):
    database = AsyncMock(DatabaseServer)
    database.database_exists.return_value = db
    database.user_exists.return_value = user
    database.user_has_all_privileges_on_database.return_value = priv
    return database


def get_spec(interval: Optional[int] = None, sec_since_last_update: Optional[int] = None) -> Spec:
    last_updated = None
    if sec_since_last_update is not None:
        last_updated = datetime.now() - timedelta(seconds=sec_since_last_update)
    return Spec(
        name="testcase",
        namespace="test",
        secret_name="secret-testcase",
        secret_user_name_postfix=None,
        password_rotation_interval=interval,
        password_last_updated=last_updated,
    )


state_observation_test_cases = [
    # Empty state
    (
        get_spec(),
        mock_db(),
        ObservedState(
            database_exists=False,
            user_exists=False,
            user_has_privileges=False,
            password_change_due=True,
        ),
    ),
    # Existing database objects, no rotation schedule
    (
        get_spec(sec_since_last_update=60 * 60),
        mock_db(db=True, user=True, priv=True),
        ObservedState(
            database_exists=True,
            user_exists=True,
            user_has_privileges=True,
            password_change_due=False,
        ),
    ),
    # Existing database objects, with last password update behind rotation schedule
    (
        get_spec(interval=60 * 60 * 24, sec_since_last_update=60 * 60 * 24 * 2),
        mock_db(db=True, user=True, priv=True),
        ObservedState(
            database_exists=True,
            user_exists=True,
            user_has_privileges=True,
            password_change_due=True,
        ),
    ),
]


@pytest.mark.parametrize("spec,db,state", state_observation_test_cases)
@pytest.mark.asyncio
async def test_observe(spec, db, state):
    controller = Controller(spec=spec, db=db)
    observed_state = await controller.observe()
    assert observed_state == state


@pytest.mark.asyncio
async def test_reconcile_creates_secret_when_password_change_due():
    state = ObservedState(
        database_exists=False,
        user_exists=False,
        user_has_privileges=False,
        password_change_due=True,
    )
    controller = Controller(db=mock_db(), spec=get_spec())
    _, secret = await controller.reconcile(state)
    assert secret is not None


@pytest.mark.asyncio
async def test_reconcile_creates_no_secret_without_password_change():
    state = ObservedState(
        database_exists=True,
        user_exists=True,
        user_has_privileges=True,
        password_change_due=False,
    )
    controller = Controller(db=mock_db(), spec=get_spec())
    _, secret = await controller.reconcile(state)
    assert secret is None


def test_status_serialization():
    status = Status().serialize()
    assert "passwordLastUpdated" in status
    assert datetime.fromisoformat(status["passwordLastUpdated"])


def test_secret_serialization(env):
    secret = Secret(
        name="secret-testcase",
        password="password",
        user="testcase@localhost",
        namespace="test",
        db="testcase",
    ).serialize()
    assert secret["data"]["PGPASSWORD"] == b64encode("password".encode()).decode()
    assert secret["data"]["PGUSER"] == b64encode("testcase@localhost".encode()).decode()
