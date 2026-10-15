from django.contrib import admin

from .models import Category, Customer, Order, Pet


@admin.register(Category)
class CategoryAdmin(admin.ModelAdmin):
    list_display = ["name"]


@admin.register(Pet)
class PetAdmin(admin.ModelAdmin):
    list_display = ["name", "price_cents", "status", "category"]
    list_filter = ["category"]
    search_fields = ["name"]
    ordering = ["id"]


@admin.register(Customer)
class CustomerAdmin(admin.ModelAdmin):
    list_display = ["name", "email"]
    search_fields = ["name", "email"]


@admin.register(Order)
class OrderAdmin(admin.ModelAdmin):
    list_display = ["pet", "customer"]
