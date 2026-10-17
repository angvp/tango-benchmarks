<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Order extends Model
{
    public $timestamps = false;

    protected $fillable = ['pet_id', 'customer_id'];

    protected $casts = ['id' => 'integer', 'pet_id' => 'integer', 'customer_id' => 'integer'];
}
