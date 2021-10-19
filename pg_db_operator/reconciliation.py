from datetime import datetime, timedelta
from typing import Optional
import string
import secrets
from dataclasses import dataclass
from loguru import logger
from .database import DatabaseServer


@dataclass
class DatabaseState:
    database_exists: bool
    user_exists: bool
    user_has_privileges: bool


@dataclass
class DatabaseSpec:
    name: str
    password_rotation_interval: Optional[int] = None

    @property
    def password(self) -> str:
        if not hasattr(self, "_password"):
            alphabet = string.ascii_letters + string.digits
            self._password = "".join(secrets.choice(alphabet) for _ in range(30))
        return self._password

    async def observe(self, db: DatabaseServer) -> DatabaseState:
        return DatabaseState(
            database_exists=await db.database_exists(self.name),
            user_exists=await db.user_exists(self.name),
            user_has_privileges=await db.user_has_all_privileges_on_database(self.name),
        )


@dataclass
class DatabaseReconciler:
    spec: DatabaseSpec
    db: DatabaseServer

    @logger.catch
    async def reconcile(self):
        state = await self.spec.observe(self.db)
        if not state.database_exists:
            await self.db.create_database(self.spec.name)
        if not state.user_exists:
            await self.db.create_user(self.spec.name, self.spec.password)
        if not state.user_has_privileges:
            await self.db.grant_all_privileges(self.spec.name)


@dataclass
class PasswordRotationReconciler:
    spec: DatabaseSpec
    db: DatabaseServer

    @logger.catch
    async def reconcile(self, password_last_updated: Optional[str] = None):
        if password_last_updated is None:
            return
        elif self.spec.password_rotation_interval is None:
            return
        else:
            now = datetime.now()
            threshold = now - timedelta(seconds=self.spec.password_rotation_interval)
            if datetime.fromisoformat(password_last_updated) < threshold:
                await self.db.update_user_password(self.spec.name, self.spec.password)
