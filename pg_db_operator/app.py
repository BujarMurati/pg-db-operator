from typing import Optional

from starlette.routing import Route
from starlette.applications import Starlette
from .database import DatabaseServer
from .health_check import health_check


def create_app(db: Optional[DatabaseServer] = None) -> Starlette:
    app = Starlette(routes=[Route("/health", endpoint=health_check)])
    app.state.db = db or DatabaseServer()
    return app
