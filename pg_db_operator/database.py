from typing import AsyncIterator
from contextlib import asynccontextmanager
from asyncpg import Connection, connect
from loguru import logger


class DatabaseServer:
    @asynccontextmanager
    async def connection(self) -> AsyncIterator[Connection]:
        connection: Connection = await connect()
        try:
            yield connection
        except Exception as e:
            logger.exception(e)
            raise e
        finally:
            await connection.close()

    async def database_exists(self, name: str) -> bool:
        async with self.connection() as conn:
            results = await conn.fetch("SELECT datname FROM pg_database WHERE datname = $1;", name)
            return len(results) > 0
