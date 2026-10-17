<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Pet extends Model
{
    public $timestamps = false;

    protected $fillable = ['name', 'price_cents', 'status', 'category_id'];

    protected $casts = ['id' => 'integer', 'price_cents' => 'integer', 'category_id' => 'integer'];
}
