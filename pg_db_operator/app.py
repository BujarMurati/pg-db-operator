from typing import Optional

from starlette.routing import Route
from starlette.applications import Starlette

from pg_db_operator.sync import sync
from .database import DatabaseServer
from .health_check import health_check


def create_app(db: Optional[DatabaseServer] = None) -> Starlette:
    app = Starlette(
        routes=[
            Route("/health", endpoint=health_check),
            Route("/sync", endpoint=sync, methods=["POST"]),
        ]
    )
    app.state.db = db or DatabaseServer()
    return app
