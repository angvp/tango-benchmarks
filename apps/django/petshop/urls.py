from django.contrib import admin
from django.urls import path

from shop import views

urlpatterns = [
    path("admin/", admin.site.urls),
    path("api/login", views.login),
    path("api/pets", views.pets),
    path("api/pets/<int:pet_id>", views.pet_detail),
    path("api/orders", views.create_order),
    path("api/orders/<int:order_id>", views.order_detail),
]
