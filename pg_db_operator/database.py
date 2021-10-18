from typing import AsyncIterator
from contextlib import asynccontextmanager
from aiopg import Connection, connect, Cursor
from loguru import logger
from psycopg2.sql import SQL, Identifier


class DatabaseServer:
    @asynccontextmanager
    async def cursor(self) -> AsyncIterator[Cursor]:
        connection: Connection = await connect()
        try:
            async with connection.cursor() as cursor:
                yield cursor
        except Exception as e:
            logger.exception(e)
            raise e
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

    async def create_database(self, name: str):
        async with self.cursor() as cursor:
            logger.info("Attempting to create database '{name}'", name=name)
            statement = SQL("CREATE DATABASE {};").format(Identifier(name))
            await cursor.execute(statement)

    async def create_user(self, name: str, password: str):
        pass
