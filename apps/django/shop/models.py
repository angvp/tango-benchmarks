from django.db import models


class Category(models.Model):
    name = models.TextField()

    def __str__(self):
        return self.name


class Pet(models.Model):
    STATUSES = [("available", "available"), ("pending", "pending"), ("sold", "sold")]

    name = models.TextField()
    price_cents = models.IntegerField()
    status = models.CharField(max_length=9, choices=STATUSES, db_index=True)
    category = models.ForeignKey(Category, on_delete=models.CASCADE)

    def __str__(self):
        return self.name


class Customer(models.Model):
    name = models.TextField()
    email = models.TextField()

    def __str__(self):
        return self.name


class Order(models.Model):
    pet = models.ForeignKey(Pet, on_delete=models.CASCADE)
    customer = models.ForeignKey(Customer, on_delete=models.CASCADE)
