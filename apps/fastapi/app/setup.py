"""Creates the tables and the staff account. Runs once, before the workers start."""
import os

from pwdlib import PasswordHash
from sqlmodel import Session, SQLModel

from app.db import engine
from app.models import Staff

SQLModel.metadata.create_all(engine)
password = os.environ.get("BENCH_STAFF_PASSWORD", "bench-staff-password")
with Session(engine) as session:
    session.merge(Staff(username="bench-staff", password_hash=PasswordHash.recommended().hash(password)))
    session.commit()
