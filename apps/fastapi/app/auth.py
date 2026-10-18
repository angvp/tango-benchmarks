import hashlib
import hmac
import os
import time

from fastapi import Depends, HTTPException
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer

LIFETIME = 12 * 60 * 60
# The entrypoint sets one secret for all workers, so a token issued by one
# worker is accepted by the other.
SECRET = os.environ["TOKEN_SECRET"].encode()
bearer = HTTPBearer(auto_error=False)


def _sign(expiry: str) -> str:
    return hmac.new(SECRET, expiry.encode(), hashlib.sha256).hexdigest()


def issue() -> str:
    expiry = str(int(time.time()) + LIFETIME)
    return f"{expiry}.{_sign(expiry)}"


def valid(token: str) -> bool:
    expiry, _, signature = token.partition(".")
    return hmac.compare_digest(signature, _sign(expiry)) and expiry.isdigit() and time.time() < int(expiry)


def require_token(credentials: HTTPAuthorizationCredentials | None = Depends(bearer)) -> None:
    if credentials is None or not valid(credentials.credentials):
        raise HTTPException(status_code=401, detail="a valid bearer token is required")
