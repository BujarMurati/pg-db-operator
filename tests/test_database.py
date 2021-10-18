import pytest
from pg_db_operator.database import DatabaseServer


@pytest.mark.asyncio
async def test_create_connection_without_error(env):
    db = DatabaseServer()
    async with db.connection():
        pass


@pytest.mark.asyncio
async def test_check_database_exists_when_database_does_not_exist(env):
    db = DatabaseServer()
    assert not await db.database_exists("i_dont_exist")


@pytest.mark.asyncio
async def test_check_database_exists_when_database_already_exists(env):
    db = DatabaseServer()
    assert await db.database_exists("database_exists")
