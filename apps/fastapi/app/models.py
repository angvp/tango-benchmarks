from sqlalchemy import CheckConstraint, Column, Integer
from sqlmodel import Field, SQLModel

STATUSES = ("available", "pending", "sold")


class Category(SQLModel, table=True):
    __tablename__ = "categories"
    id: int | None = Field(default=None, primary_key=True)
    name: str


class Pet(SQLModel, table=True):
    __tablename__ = "pets"
    __table_args__ = (
        CheckConstraint("price_cents >= 0"),
        CheckConstraint("status IN ('available', 'pending', 'sold')"),
    )
    id: int | None = Field(default=None, primary_key=True)
    name: str
    price_cents: int = Field(sa_column=Column(Integer, nullable=False))
    status: str = Field(index=True)
    category_id: int = Field(foreign_key="categories.id", index=True)


class Customer(SQLModel, table=True):
    __tablename__ = "customers"
    id: int | None = Field(default=None, primary_key=True)
    name: str
    email: str


class Order(SQLModel, table=True):
    __tablename__ = "orders"
    id: int | None = Field(default=None, primary_key=True)
    pet_id: int = Field(foreign_key="pets.id")
    customer_id: int = Field(foreign_key="customers.id")


class Staff(SQLModel, table=True):
    __tablename__ = "staff"
    username: str = Field(primary_key=True)
    password_hash: str
