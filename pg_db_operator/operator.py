import secrets
import string
from typing import Dict, Optional
from base64 import b64encode
import kopf
from dotenv import load_dotenv
from yaml import safe_load
from kubernetes import client
from pg_db_operator.database import DatabaseServer
from pg_db_operator.models import PostgresDatabaseSpec

load_dotenv()

db = DatabaseServer()

secret_template = """
apiVersion: v1
metadata:
    name: {name}
data:
  PGUSER: {user}
  PGPASSWORD: {password}
  PGDATABASE: {database}
kind: Secret
type: Opaque
"""


def _generate_password():
    alphabet = string.ascii_letters + string.digits
    return "".join(secrets.choice(alphabet) for _ in range(30))


async def upsert_database_and_user(name: str, logger: kopf.Logger) -> Optional[str]:
    password = None
    if not await db.database_exists(name):
        await db.create_database(name)
        logger.info(f"Created database {name}")
    if not await db.user_exists(name):
        password = _generate_password()
        await db.create_user(name, password)
        logger.info(f"Created user {name}")
    if not await db.user_has_all_privileges_on_database(name):
        await db.grant_all_privileges(name)
    return password


def _encode_secret_data(value: str) -> str:
    return b64encode(value.encode()).decode()


@kopf.on.create("postgresdatabases")
async def on_create(spec: Dict, namespace: str, logger, **_):
    db_spec = PostgresDatabaseSpec(**spec)
    password = await upsert_database_and_user(db_spec.name, logger)
    if password is None:
        password = _generate_password()
        await db.update_user_password(db_spec.name, password)
    api = client.CoreV1Api()
    secret_dict = safe_load(
        secret_template.format(
            name=db_spec.target_secret.name,
            user=_encode_secret_data(db_spec.user),
            password=_encode_secret_data(password),
            database=_encode_secret_data(db_spec.name),
        )
    )
    api.create_namespaced_secret(namespace=namespace, body=secret_dict)
    logger.info(f"created secret {db_spec.target_secret.name} in namespace {namespace}")


@kopf.on.update("postgresdatabases")
async def on_update(spec, old, namespace, logger, **_):
    old_spec = PostgresDatabaseSpec(**old["spec"])
    db_spec = PostgresDatabaseSpec(**spec)
    password = await upsert_database_and_user(db_spec.name, logger)
    if password is None:
        password = _generate_password()
        await db.update_user_password(db_spec.name, password)
    api = client.CoreV1Api()
    api.delete_namespaced_secret(namespace=namespace, name=old_spec.target_secret.name)
    logger.info(f"deleted secret {old_spec.target_secret.name} from namespace {namespace}")
    secret_dict = safe_load(
        secret_template.format(
            name=db_spec.target_secret.name,
            user=_encode_secret_data(db_spec.user),
            password=_encode_secret_data(password),
            database=_encode_secret_data(db_spec.name),
        )
    )
    api.create_namespaced_secret(namespace=namespace, body=secret_dict)
    logger.info(f"recreated secret {db_spec.target_secret.name} in namespace {namespace}")


@kopf.on.delete("postgresdatabases")
async def delete(spec, namespace, logger, **_):
    db_spec = PostgresDatabaseSpec(**spec)
    api = client.CoreV1Api()
    api.delete_namespaced_secret(namespace=namespace, name=db_spec.target_secret.name)
    logger.info(f"deleted secret {db_spec.target_secret.name} from namespace {namespace}")
