import pytest
from psycopg2.sql import SQL, Identifier
from pg_db_operator.database import DatabaseServer


class TestDatabaseServer(DatabaseServer):
    def __init__(self):
        self.clean_up = set()

    async def create_database(self, name: str):
        self.clean_up.add(name)
        await super().create_database(name)

    async def create_user(self, name: str, password: str):
        self.clean_up.add(name)
        await super().create_user(name, password)

    async def cleanup(self):
        for name in self.clean_up:
            async with self.cursor() as cursor:
                if await self.user_has_all_privileges_on_database(name):
                    await cursor.execute(
                        SQL("REVOKE ALL PRIVILEGES ON DATABASE {} FROM {};").format(
                            Identifier(name), Identifier(name)
                        )
                    )
                await cursor.execute(SQL("DROP USER IF EXISTS {};").format(Identifier(name)))
                await cursor.execute(SQL("DROP DATABASE IF EXISTS {};").format(Identifier(name)))


@pytest.fixture
async def db(env):
    db = TestDatabaseServer()
    try:
        yield db
    finally:
        await db.cleanup()


@pytest.mark.asyncio
async def test_create_connection_without_error(env):
    db = DatabaseServer()
    async with db.cursor():
        pass


@pytest.mark.asyncio
async def test_check_database_exists_when_database_does_not_exist(env):
    db = DatabaseServer()
    assert not await db.database_exists("i_dont_exist")


@pytest.mark.asyncio
async def test_check_database_exists_when_database_already_exists(env):
    db = DatabaseServer()
    assert await db.database_exists("database_exists")


@pytest.mark.asyncio
async def test_check_user_exists_when_user_does_not_exist(env):
    db = DatabaseServer()
    assert not await db.user_exists("i_dont_exist")


@pytest.mark.asyncio
async def test_check_user_exists_when_user_already_exists(env):
    db = DatabaseServer()
    assert await db.user_exists("user_exists")


@pytest.mark.asyncio
async def test_check_user_has_all_privileges_when_user_has_all_privileges(db):
    assert await db.user_has_all_privileges_on_database("everything_exists")


@pytest.mark.asyncio
async def test_check_user_has_all_privileges_when_user_no_privileges(db):
    assert not await db.user_has_all_privileges_on_database("user_and_db_but_no_privileged")


@pytest.mark.asyncio
async def test_create_database(db):
    name = "new_database"
    await db.create_database(name)
    assert await db.database_exists(name)


@pytest.mark.asyncio
async def test_create_user(db):
    name = "new_database"
    password = "test"
    await db.create_user(name, password)
    assert await db.user_exists(name)


@pytest.mark.asyncio
async def test_grant_privileges(db):
    name = "new_database"
    password = "test"
    await db.create_user(name, password)
    await db.create_database(name)
    await db.grant_all_privileges(name)
    assert await db.user_has_all_privileges_on_database(name)
