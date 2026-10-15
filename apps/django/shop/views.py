import json

from django.contrib.auth import authenticate
from django.core import signing
from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods

from .models import Category, Customer, Order, Pet

TOKEN_SALT = "petshop.api"
TOKEN_MAX_AGE = 12 * 60 * 60
DEFAULT_PAGE_SIZE = 20
MAX_PAGE_SIZE = 100
STATUSES = {"available", "pending", "sold"}


def fail(status, message):
    return JsonResponse({"error": message}, status=status)


def pet_json(pet):
    return {
        "id": pet.id,
        "name": pet.name,
        "price_cents": pet.price_cents,
        "status": pet.status,
        "category_id": pet.category_id,
    }


def order_json(order):
    return {"id": order.id, "pet_id": order.pet_id, "customer_id": order.customer_id}


def read_body(request):
    try:
        body = json.loads(request.body or b"{}")
    except ValueError:
        return None
    return body if isinstance(body, dict) else None


def authorised(request):
    header = request.headers.get("Authorization", "")
    if not header.startswith("Bearer "):
        return False
    try:
        signing.loads(header[7:], salt=TOKEN_SALT, max_age=TOKEN_MAX_AGE)
    except signing.BadSignature:
        return False
    return True


def token_required(view):
    def wrapped(request, *args, **kwargs):
        if not authorised(request):
            return fail(401, "a valid bearer token is required")
        return view(request, *args, **kwargs)

    return wrapped


@csrf_exempt
@require_http_methods(["POST"])
def login(request):
    body = read_body(request)
    if body is None:
        return fail(400, "body is not valid JSON")
    user = authenticate(request, username=body.get("username"), password=body.get("password"))
    if user is None or not user.is_staff:
        return fail(401, "bad credentials")
    return JsonResponse({"token": signing.dumps({"user": user.pk}, salt=TOKEN_SALT)})


def int_param(request, name):
    """Returns (value, present, ok) for an optional integer query parameter."""
    raw = request.GET.get(name)
    if raw is None or raw == "":
        return 0, False, True
    try:
        return int(raw), True, True
    except ValueError:
        return 0, True, False


@csrf_exempt
@require_http_methods(["GET", "POST"])
def pets(request):
    if request.method == "POST":
        return token_required(create_pet)(request)
    return list_pets(request)


def list_pets(request):
    category, has_category, ok1 = int_param(request, "category")
    min_price, has_min, ok2 = int_param(request, "min_price")
    max_price, has_max, ok3 = int_param(request, "max_price")
    page, has_page, ok4 = int_param(request, "page")
    size, has_size, ok5 = int_param(request, "page_size")
    if not (ok1 and ok2 and ok3 and ok4 and ok5):
        return fail(400, "numeric parameters must be integers")
    page = page if has_page else 1
    size = size if has_size else DEFAULT_PAGE_SIZE
    if page < 1 or size < 1 or size > MAX_PAGE_SIZE:
        return fail(400, "page must be at least 1 and page_size between 1 and 100")
    status = request.GET.get("status", "")
    if status and status not in STATUSES:
        return fail(400, "unknown status")
    queryset = Pet.objects.all()
    if has_category:
        queryset = queryset.filter(category_id=category)
    if status:
        queryset = queryset.filter(status=status)
    if has_min:
        queryset = queryset.filter(price_cents__gte=min_price)
    if has_max:
        queryset = queryset.filter(price_cents__lte=max_price)
    total = queryset.count()
    start = (page - 1) * size
    items = [pet_json(p) for p in queryset.order_by("id")[start:start + size]]
    return JsonResponse({"items": items, "page": page, "page_size": size, "total": total})


@csrf_exempt
@require_http_methods(["GET", "PUT"])
def pet_detail(request, pet_id):
    if request.method == "PUT":
        return token_required(update_pet)(request, pet_id)
    pet = Pet.objects.filter(pk=pet_id).first()
    if pet is None:
        return fail(404, "no such pet")
    return JsonResponse(pet_json(pet))


def pet_fields(body):
    """Returns (fields, problem): the validated fields, or what is wrong."""
    name, price, status, category = body.get("name"), body.get("price_cents"), body.get("status"), body.get("category_id")
    if not isinstance(name, str) or not name:
        return None, "name is required"
    if not isinstance(price, int) or isinstance(price, bool) or price < 0 or price > 2147483647:
        return None, "price_cents must be between 0 and 2147483647"
    if status not in STATUSES:
        return None, "status must be available, pending or sold"
    if not isinstance(category, int) or isinstance(category, bool) or not Category.objects.filter(pk=category).exists():
        return None, "category_id names no category"
    return {"name": name, "price_cents": price, "status": status, "category_id": category}, None


def create_pet(request):
    body = read_body(request)
    if body is None:
        return fail(400, "body is not valid JSON")
    fields, problem = pet_fields(body)
    if problem:
        return fail(422, problem)
    return JsonResponse(pet_json(Pet.objects.create(**fields)), status=201)


def update_pet(request, pet_id):
    pet = Pet.objects.filter(pk=pet_id).first()
    if pet is None:
        return fail(404, "no such pet")
    body = read_body(request)
    if body is None:
        return fail(400, "body is not valid JSON")
    fields, problem = pet_fields(body)
    if problem:
        return fail(422, problem)
    for key, value in fields.items():
        setattr(pet, key, value)
    pet.save()
    return JsonResponse(pet_json(pet))


@csrf_exempt
@require_http_methods(["POST"])
@token_required
def create_order(request):
    body = read_body(request)
    if body is None:
        return fail(400, "body is not valid JSON")
    pet_id, customer_id = body.get("pet_id"), body.get("customer_id")
    valid = all(isinstance(v, int) and not isinstance(v, bool) for v in (pet_id, customer_id))
    if not valid or not Pet.objects.filter(pk=pet_id).exists() or not Customer.objects.filter(pk=customer_id).exists():
        return fail(422, "pet_id and customer_id must name existing rows")
    return JsonResponse(order_json(Order.objects.create(pet_id=pet_id, customer_id=customer_id)), status=201)


@csrf_exempt
@require_http_methods(["GET"])
@token_required
def order_detail(request, order_id):
    order = Order.objects.filter(pk=order_id).first()
    if order is None:
        return fail(404, "no such order")
    return JsonResponse(order_json(order))
