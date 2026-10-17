<?php

namespace App\Http\Controllers;

use App\Models\User;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Hash;

class SessionController extends Controller
{
    public function store(Request $request): JsonResponse
    {
        $user = User::where('name', (string) $request->input('username'))->first();
        if (! $user || ! Hash::check((string) $request->input('password'), $user->password)) {
            return response()->json(['error' => 'bad credentials'], 401);
        }

        return response()->json(['token' => $user->createToken('api')->plainTextToken]);
    }
}
