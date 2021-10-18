from starlette.exceptions import HTTPException
from starlette.responses import JSONResponse
from starlette.requests import Request
from loguru import logger


async def health_check(request: Request) -> JSONResponse:
    try:
        async with request.app.state.db.connection():
            pass
        return JSONResponse({"status": "pass"})
    except Exception as e:
        logger.exception("Failed to connect")
        raise HTTPException(500, detail="Could not connect to database server") from e
