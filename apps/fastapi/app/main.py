from typing import Annotated, Any

from fastapi import Body, Depends, FastAPI, HTTPException, Query, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from pwdlib import PasswordHash
from sqlalchemy import func
from sqlmodel import Session, select
from starlette.exceptions import HTTPException as StarletteHTTPException

from app import auth
from app.db import get_session
from app.models import STATUSES, Category, Customer, Order, Pet, Staff

app = FastAPI()
hasher = PasswordHash.recommended()
SessionDep = Annotated[Session, Depends(get_session)]


# The contract's errors are {"error": "..."}; FastAPI's are {"detail": ...}.
@app.exception_handler(StarletteHTTPException)
async def http_error(_: Request, exc: StarletteHTTPException) -> JSONResponse:
    return JSONResponse({"error": str(exc.detail)}, status_code=exc.status_code)


@app.exception_handler(RequestValidationError)
async def invalid_request(_: Request, __: RequestValidationError) -> JSONResponse:
    return JSONResponse({"error": "the request is not valid"}, status_code=400)


def pet_json(pet: Pet) -> dict[str, Any]:
    return pet.model_dump()


def to_int(value: Any) -> int | None:
    return value if isinstance(value, int) and not isinstance(value, bool) else None


def pet_problem(body: dict[str, Any], session: Session) -> str | None:
    """What is wrong with a pet body, or None. The contract's 422 rules are
    written here; the body is read as a plain dict so that auth and the 404
    come first."""
    price, category = to_int(body.get("price_cents")), to_int(body.get("category_id"))
    if not isinstance(body.get("name"), str) or not body["name"]:
        return "name is required"
    if price is None or price < 0 or price > 2147483647:
        return "price_cents must be between 0 and 2147483647"
    if body.get("status") not in STATUSES:
        return "status must be available, pending or sold"
    if category is None or session.get(Category, category) is None:
        return "category_id names no category"
    return None


@app.post("/api/login")
def login(session: SessionDep, body: Annotated[dict[str, Any], Body()]) -> dict[str, str]:
    staff = session.get(Staff, str(body.get("username")))
    if staff is None or not hasher.verify(str(body.get("password")), staff.password_hash):
        raise HTTPException(status_code=401, detail="bad credentials")
    return {"token": auth.issue()}


@app.get("/api/pets")
def list_pets(
    session: SessionDep,
    category: int | None = None,
    status: str | None = None,
    min_price: int | None = None,
    max_price: int | None = None,
    page: Annotated[int, Query(ge=1)] = 1,
    page_size: Annotated[int, Query(ge=1, le=100)] = 20,
) -> dict[str, Any]:
    if status is not None and status not in STATUSES:
        raise HTTPException(status_code=400, detail="unknown status")
    conditions = []
    if category is not None:
        conditions.append(Pet.category_id == category)
    if status is not None:
        conditions.append(Pet.status == status)
    if min_price is not None:
        conditions.append(Pet.price_cents >= min_price)
    if max_price is not None:
        conditions.append(Pet.price_cents <= max_price)
    total = session.exec(select(func.count()).select_from(Pet).where(*conditions)).one()
    pets = session.exec(select(Pet).where(*conditions).order_by(Pet.id).offset((page - 1) * page_size).limit(page_size)).all()
    return {"items": [pet_json(p) for p in pets], "page": page, "page_size": page_size, "total": total}


def find(session: Session, model: Any, raw_id: str, what: str) -> Any:
    found = session.get(model, int(raw_id)) if raw_id.isdigit() and 0 < int(raw_id) <= 2147483647 else None
    if found is None:
        raise HTTPException(status_code=404, detail=f"no such {what}")
    return found


@app.get("/api/pets/{pet_id}")
def get_pet(pet_id: str, session: SessionDep) -> dict[str, Any]:
    return pet_json(find(session, Pet, pet_id, "pet"))


@app.post("/api/pets", status_code=201, dependencies=[Depends(auth.require_token)])
def create_pet(session: SessionDep, body: Annotated[dict[str, Any], Body()]) -> dict[str, Any]:
    problem = pet_problem(body, session)
    if problem:
        raise HTTPException(status_code=422, detail=problem)
    pet = Pet(name=body["name"], price_cents=body["price_cents"], status=body["status"], category_id=body["category_id"])
    session.add(pet)
    session.commit()
    session.refresh(pet)
    return pet_json(pet)


@app.put("/api/pets/{pet_id}", dependencies=[Depends(auth.require_token)])
def update_pet(pet_id: str, session: SessionDep, body: Annotated[dict[str, Any], Body()]) -> dict[str, Any]:
    pet = find(session, Pet, pet_id, "pet")
    problem = pet_problem(body, session)
    if problem:
        raise HTTPException(status_code=422, detail=problem)
    pet.name, pet.price_cents, pet.status, pet.category_id = body["name"], body["price_cents"], body["status"], body["category_id"]
    session.add(pet)
    session.commit()
    session.refresh(pet)
    return pet_json(pet)


@app.post("/api/orders", status_code=201, dependencies=[Depends(auth.require_token)])
def create_order(session: SessionDep, body: Annotated[dict[str, Any], Body()]) -> dict[str, Any]:
    pet_id, customer_id = to_int(body.get("pet_id")), to_int(body.get("customer_id"))
    if pet_id is None or customer_id is None or session.get(Pet, pet_id) is None or session.get(Customer, customer_id) is None:
        raise HTTPException(status_code=422, detail="pet_id and customer_id must name existing rows")
    order = Order(pet_id=pet_id, customer_id=customer_id)
    session.add(order)
    session.commit()
    session.refresh(order)
    return order.model_dump()


@app.get("/api/orders/{order_id}", dependencies=[Depends(auth.require_token)])
def get_order(order_id: str, session: SessionDep) -> dict[str, Any]:
    return find(session, Order, order_id, "order").model_dump()
