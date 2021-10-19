import pytest
from datetime import datetime, timedelta
from unittest.mock import AsyncMock
from pg_db_operator.database import DatabaseServer
from pg_db_operator.reconciliation import (
    DatabaseSpec,
    DatabaseReconciler,
    PasswordRotationReconciler,
)


@pytest.fixture
def db():
    return AsyncMock(DatabaseServer)


@pytest.mark.asyncio
async def test_database_spec_observe(db):
    db.database_exists.return_value = True
    db.user_exists.return_value = True
    db.user_has_all_privileges_on_database.return_value = True
    spec = DatabaseSpec(name="test")
    observed_state = await spec.observe(db)
    assert observed_state.database_exists
    assert observed_state.user_exists
    assert observed_state.user_has_privileges


@pytest.mark.asyncio
async def test_db_reconciler_creates_everything_when_nothing_exists(db):
    db.database_exists.return_value = False
    db.user_exists.return_value = False
    db.user_has_all_privileges_on_database.return_value = False

    reconciler = DatabaseReconciler(spec=DatabaseSpec(name="test"), db=db)
    await reconciler.reconcile()
    db.create_database.assert_awaited_with("test")
    db.create_user.assert_awaited()
    assert db.create_user.await_args[0][0] == "test"
    assert type(db.create_user.await_args[0][1]) == str
    assert len(db.create_user.await_args[0][1]) == 30
    db.grant_all_privileges.assert_awaited_with("test")


@pytest.mark.asyncio
async def test_db_reconciler_skips_creating_db_if_exists(db):
    db.database_exists.return_value = True
    db.user_exists.return_value = False
    db.user_has_all_privileges_on_database.return_value = False

    reconciler = DatabaseReconciler(spec=DatabaseSpec(name="test"), db=db)
    await reconciler.reconcile()
    db.create_database.assert_not_awaited()
    db.create_user.assert_awaited()
    assert db.create_user.await_args[0][0] == "test"
    assert type(db.create_user.await_args[0][1]) == str
    assert len(db.create_user.await_args[0][1]) == 30
    db.grant_all_privileges.assert_awaited_with("test")


@pytest.mark.asyncio
async def test_db_reconciler_skips_creating_user_if_exists(db):
    db.database_exists.return_value = False
    db.user_exists.return_value = True
    db.user_has_all_privileges_on_database.return_value = False

    reconciler = DatabaseReconciler(spec=DatabaseSpec(name="test"), db=db)
    await reconciler.reconcile()
    db.create_database.assert_awaited_with("test")
    db.create_user.assert_not_awaited()
    db.grant_all_privileges.assert_awaited_with("test")


@pytest.mark.asyncio
async def test_pw_reconciler_skips_password_update_if_no_interval_is_set(db):
    reconciler = PasswordRotationReconciler(
        spec=DatabaseSpec(name="test"),
        db=db,
    )
    await reconciler.reconcile()
    db.update_user_password.assert_not_awaited()


@pytest.mark.asyncio
async def test_pw_reconciler_skips_password_update_if_never_updated(db):
    reconciler = PasswordRotationReconciler(
        spec=DatabaseSpec(name="test", password_rotation_interval=60 * 60 * 24 * 7),
        db=db,
    )
    await reconciler.reconcile()
    db.update_user_password.assert_not_awaited()


@pytest.mark.asyncio
async def test_pw_reconciler_skips_password_update_if_recently_updated(db):
    weekly_interval = 60 * 60 * 24 * 7
    reconciler = PasswordRotationReconciler(
        spec=DatabaseSpec(name="test", password_rotation_interval=weekly_interval),
        db=db,
    )
    updated_yesterday = datetime.now() - timedelta(seconds=weekly_interval / 7)
    await reconciler.reconcile(password_last_updated=updated_yesterday.isoformat())
    db.update_user_password.assert_not_awaited()


@pytest.mark.asyncio
async def test_pw_reconciler_updates_password_if_rotation_due(db):
    weekly_interval = 60 * 60 * 24 * 7
    reconciler = PasswordRotationReconciler(
        spec=DatabaseSpec(name="test", password_rotation_interval=weekly_interval),
        db=db,
    )
    updated_one_and_a_half_weeks_ago = datetime.now() - timedelta(seconds=1.5 * weekly_interval)
    await reconciler.reconcile(password_last_updated=updated_one_and_a_half_weeks_ago.isoformat())
    db.update_user_password.assert_awaited()
