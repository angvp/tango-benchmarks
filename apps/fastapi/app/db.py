import os
from collections.abc import Iterator

from sqlmodel import Session, create_engine

engine = create_engine(os.environ["DATABASE_URL"])


def get_session() -> Iterator[Session]:
    with Session(engine) as session:
        yield session
