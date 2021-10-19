from dotenv import load_dotenv
import uvicorn
from pg_db_operator.app import create_app

load_dotenv()

app = create_app()

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)  # type: ignore
