import secrets
from base64 import b64encode
import os
from yaml import safe_load
import string
from typing import Dict, Optional, Tuple
from datetime import datetime, timedelta
from dataclasses import dataclass, field

from pg_db_operator.database import DatabaseServer

SECRET_TEMPLATE = """
apiVersion: v1
data:
  PGUSER: {user}
  PGPASSWORD: {password}
  PGHOST: {host}
  PGDATABASE: {db}
kind: Secret
metadata:
  name: {name}
  namespace: {namespace}
type: Opaque
"""


@dataclass
class ObservedState:
    database_exists: bool
    user_exists: bool
    user_has_privileges: bool
    password_change_due: bool


@dataclass
class Spec:
    name: str
    namespace: str
    secret_name: str
    secret_user_name_postfix: Optional[str] = None
    password_rotation_interval: Optional[int] = None
    password_last_updated: Optional[datetime] = None


@dataclass
class Secret:
    name: str
    namespace: str
    password: str
    user: str
    db: str

    def _encode(self, value: str) -> str:
        return b64encode(value.encode()).decode()

    def serialize(self) -> Dict:
        return safe_load(
            SECRET_TEMPLATE.format(
                name=self.name,
                namespace=self.namespace,
                host=self._encode(os.environ["PGHOST"]),
                port=self._encode(os.environ["PGPORT"]),
                password=self._encode(self.password),
                db=self._encode(self.db),
                user=self._encode(self.user),
            )
        )


@dataclass
class Status:
    password_last_updated: datetime = field(default_factory=datetime.now)

    def serialize(self) -> Dict[str, str]:
        return {"passwordLastUpdated": self.password_last_updated.isoformat()}


@dataclass
class Controller:
    spec: Spec
    db: DatabaseServer

    def _generate_password(self) -> str:
        alphabet = string.ascii_letters + string.digits
        return "".join(secrets.choice(alphabet) for _ in range(30))

    def _evaluate_password_rotation_requirement(self) -> bool:
        """
        Determine if a password rotation is in order.
        If no password rotation interval is defined, this will always return False.
        It will also return false if the password has never been updated before (e.g.
        initial creation)
        """
        if self.spec.password_last_updated is None:
            return False
        elif self.spec.password_rotation_interval is None:
            return False
        else:
            now = datetime.now()
            threshold = now - timedelta(seconds=self.spec.password_rotation_interval)
            return self.spec.password_last_updated < threshold

    async def observe(self) -> ObservedState:
        database_exists = await self.db.database_exists(self.spec.name)
        user_exists = await self.db.user_exists(self.spec.name)
        user_has_privileges = await self.db.user_has_all_privileges_on_database(self.spec.name)
        password_change_due = not user_exists or self._evaluate_password_rotation_requirement()
        return ObservedState(
            database_exists=database_exists,
            user_exists=user_exists,
            user_has_privileges=user_has_privileges,
            password_change_due=password_change_due,
        )

    async def reconcile(self, state: ObservedState) -> Tuple[Status, Optional[Secret]]:
        secret = None
        password = self._generate_password()
        if state.password_change_due:
            user = self.spec.name + (self.spec.secret_user_name_postfix or "")
            secret = Secret(
                name=self.spec.secret_name,
                user=user,
                password=password,
                namespace=self.spec.namespace,
                db=self.spec.name,
            )
        if not state.database_exists:
            await self.db.create_database(self.spec.name)
        if not state.user_exists:
            await self.db.create_user(self.spec.name, password)
        if not state.user_has_privileges:
            await self.db.grant_all_privileges(self.spec.name)
        if state.password_change_due and state.user_exists:
            await self.db.update_user_password(self.spec.name, password)
        return Status(), secret
