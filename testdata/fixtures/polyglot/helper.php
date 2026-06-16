<?php
namespace App;

use App\Support\Str;

class Helper {
    public function clean($value) {
        return normalize($value);
    }
}

function normalize($v) {
    return Str::lower($v);
}
