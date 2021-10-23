import pytest
from psycopg2.sql import SQL, Identifier
from pg_db_operator.database import DatabaseServer


@pytest.fixture
def env(monkeypatch) -> None:
    """Set environment variables for connecting to the database"""
    monkeypatch.setenv("PGHOST", "localhost")
    monkeypatch.setenv("PGPASSWORD", "test")
    monkeypatch.setenv("PGUSER", "server_admin")
    monkeypatch.setenv("PGPORT", "5432")
    monkeypatch.setenv("PGDATABASE", "postgres")


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
