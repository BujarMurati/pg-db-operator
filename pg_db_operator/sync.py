from datetime import datetime
from starlette.requests import Request
from starlette.responses import JSONResponse
from .controller import Controller, Spec


async def sync(request: Request) -> JSONResponse:
    body = await request.json()
    parent = body["parent"]
    password_last_updated = parent.get("status", {}).get("passwordLastUpdated")
    if password_last_updated is not None:
        password_last_updated = datetime.fromisoformat(password_last_updated)
    spec = Spec(
        name=parent["spec"]["name"],
        namespace=parent["metadata"]["namespace"],
        secret_name=parent["spec"]["targetSecret"]["name"],
        secret_user_name_postfix=parent["spec"]["targetSecret"].get("userNamePostfix"),
        password_rotation_interval=parent["spec"].get("passwordRotationInterval"),
        password_last_updated=password_last_updated,
    )
    controller = Controller(db=request.app.state.db, spec=spec)
    state = await controller.observe()
    status, secret = await controller.reconcile(state)
    if secret is not None:
        response = {
            "status": status.serialize(),
            "children": [secret.serialize()],
        }
    else:
        response = {
            "status": status.serialize(),
            "children": [body["children"]["Secret.v1"].values()],
        }
    return JSONResponse(response)
