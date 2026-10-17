<?php

namespace App\Http\Controllers;

use App\Models\Order;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Validation\Rule;

class OrderController extends Controller
{
    public function store(Request $request): JsonResponse
    {
        $validator = validator($request->json()->all(), [
            'pet_id' => ['required', 'integer', 'min:1', 'max:2147483647', Rule::exists('pets', 'id')],
            'customer_id' => ['required', 'integer', 'min:1', 'max:2147483647', Rule::exists('customers', 'id')],
        ]);
        if ($validator->fails()) {
            return response()->json(['error' => 'pet_id and customer_id must name existing rows'], 422);
        }

        return response()->json(Order::create($validator->validated())->refresh(), 201);
    }

    public function show(string $id): JsonResponse
    {
        $order = ctype_digit($id) && (int) $id <= 2147483647 ? Order::find((int) $id) : null;

        return $order ? response()->json($order) : response()->json(['error' => 'no such order'], 404);
    }
}
