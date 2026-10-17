<?php

namespace App\Http\Controllers;

use App\Models\Category;
use App\Models\Pet;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Validation\Rule;

class PetController extends Controller
{
    private const MAX_INT32 = 2147483647;

    public function index(Request $request): JsonResponse
    {
        $input = $request->query();
        $numbers = [];
        foreach (['category', 'min_price', 'max_price', 'page', 'page_size'] as $name) {
            $raw = $input[$name] ?? '';
            if ($raw === '') {
                $numbers[$name] = null;
                continue;
            }
            if (! is_string($raw) || ! preg_match('/^[+-]?\d+$/', $raw) || abs((int) $raw) > self::MAX_INT32) {
                return response()->json(['error' => 'numeric parameters must be integers'], 400);
            }
            $numbers[$name] = (int) $raw;
        }
        $page = $numbers['page'] ?? 1;
        $size = $numbers['page_size'] ?? 20;
        if ($page < 1 || $size < 1 || $size > 100) {
            return response()->json(['error' => 'page must be at least 1 and page_size between 1 and 100'], 400);
        }
        $status = $input['status'] ?? '';
        if ($status !== '' && ! in_array($status, ['available', 'pending', 'sold'], true)) {
            return response()->json(['error' => 'unknown status'], 400);
        }

        $query = Pet::query();
        if ($numbers['category'] !== null) {
            $query->where('category_id', $numbers['category']);
        }
        if ($status !== '') {
            $query->where('status', $status);
        }
        if ($numbers['min_price'] !== null) {
            $query->where('price_cents', '>=', $numbers['min_price']);
        }
        if ($numbers['max_price'] !== null) {
            $query->where('price_cents', '<=', $numbers['max_price']);
        }

        $total = $query->count();

        return response()->json([
            'items' => $query->orderBy('id')->offset(($page - 1) * $size)->limit($size)->get(),
            'page' => $page,
            'page_size' => $size,
            'total' => $total,
        ]);
    }

    public function show(string $id): JsonResponse
    {
        $pet = ctype_digit($id) && (int) $id <= self::MAX_INT32 ? Pet::find((int) $id) : null;

        return $pet ? response()->json($pet) : response()->json(['error' => 'no such pet'], 404);
    }

    public function store(Request $request): JsonResponse
    {
        $data = $this->validated($request);
        if ($data instanceof JsonResponse) {
            return $data;
        }

        return response()->json(Pet::create($data)->refresh(), 201);
    }

    public function update(Request $request, string $id): JsonResponse
    {
        $pet = ctype_digit($id) && (int) $id <= self::MAX_INT32 ? Pet::find((int) $id) : null;
        if (! $pet) {
            return response()->json(['error' => 'no such pet'], 404);
        }
        $data = $this->validated($request);
        if ($data instanceof JsonResponse) {
            return $data;
        }
        $pet->update($data);

        return response()->json($pet);
    }

    /** The four fields of a pet, or the 422 response that says what is wrong. */
    private function validated(Request $request): array|JsonResponse
    {
        $validator = validator($request->json()->all(), [
            'name' => ['required', 'string'],
            'price_cents' => ['required', 'integer', 'min:0', 'max:'.self::MAX_INT32],
            'status' => ['required', Rule::in(['available', 'pending', 'sold'])],
            'category_id' => ['required', 'integer', 'min:1', 'max:'.self::MAX_INT32, Rule::exists(Category::class, 'id')],
        ]);
        if ($validator->fails()) {
            return response()->json(['error' => $validator->errors()->first()], 422);
        }

        return $validator->validated();
    }
}
