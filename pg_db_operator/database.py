from typing import AsyncIterator
from contextlib import asynccontextmanager
from aiopg import Connection, connect, Cursor
from psycopg2.sql import SQL, Identifier


class DatabaseServer:
    @asynccontextmanager
    async def cursor(self) -> AsyncIterator[Cursor]:
        connection: Connection = await connect()
        try:
            async with connection.cursor() as cursor:
                yield cursor
        finally:
            await connection.close()

    async def database_exists(self, name: str) -> bool:
        async with self.cursor() as cursor:
            await cursor.execute("SELECT datname FROM pg_database WHERE datname = %s;", (name,))
            results = await cursor.fetchall()
            return len(results) > 0

    async def user_exists(self, name: str) -> bool:
        async with self.cursor() as cursor:
            await cursor.execute("SELECT usename FROM pg_user WHERE usename = %s;", (name,))
            results = await cursor.fetchall()
            return len(results) > 0

    async def user_has_all_privileges_on_database(self, name: str) -> bool:
        async with self.cursor() as cursor:
            if not (await self.database_exists(name) and await self.user_exists(name)):
                return False
            # https://www.postgresql.org/docs/current/functions-info.html#FUNCTIONS-INFO-ACCESS-TABLE
            # Here we test for CREATE permissions, there does not seem to be an elegant way to
            # verify that "ALL" privileges on a database were granted
            await cursor.execute("SELECT has_database_privilege(%s, %s, 'CREATE')", (name, name))
            result = await cursor.fetchone()
            return result[0]

    async def create_database(self, name: str):
        async with self.cursor() as cursor:
            statement = SQL("CREATE DATABASE {};").format(Identifier(name))
            await cursor.execute(statement)

    async def create_user(self, name: str, password: str):
        async with self.cursor() as cursor:
            statement = SQL("CREATE USER {} WITH ENCRYPTED PASSWORD %s;").format(Identifier(name))
            await cursor.execute(statement, (password,))

    async def grant_all_privileges(self, name: str):
        async with self.cursor() as cursor:
            statement = SQL("GRANT ALL PRIVILEGES ON DATABASE {} TO {}").format(
                Identifier(name), Identifier(name)
            )
            await cursor.execute(statement)

    async def update_user_password(self, name: str, password: str):
        async with self.cursor() as cursor:
            statement = SQL("ALTER USER {} WITH ENCRYPTED PASSWORD %s;").format(Identifier(name))
            await cursor.execute(statement, (password,))
