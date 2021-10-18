import pytest
from unittest.mock import AsyncMock
from pg_db_operator.database import DatabaseServer
from pg_db_operator.reconciliation import DatabaseSpec, DatabaseReconciler


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
