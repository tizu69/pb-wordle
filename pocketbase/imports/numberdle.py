#!/usr/bin/env python
# This script put intdle.json and floatdle.json into PWD.
import json
import random


def put(name, words):
    with open(name.lower() + ".json", "w") as f:
        json.dump({"name": name, "mustHint": False, "mustPresent": False, "words": words}, f)


def count_decimal_places(n):
    str_num = str(n)
    if "." in str_num:
        nr = len(str_num.split(".")[1])
        return f"{nr} decimal place{'' if nr == 1 else 's'}"


int_numbers = [random.randint(0, 1000000) for _ in range(10000)]
put("Intle", [{"word": str(n)} for n in int_numbers])

float_numbers = []
for _ in range(10000):
    num = random.randint(0, 1000000)
    num_str = str(num)
    if len(num_str) == 1:
        # single digit numbers should have a leading 0
        float_num = float("0." + num_str)
    else:
        pos = random.randint(1, len(num_str))
        float_str = num_str[:pos] + "." + num_str[pos:]
        float_num = float(float_str)
    float_numbers.append(float_num)
float_numbers_mapped = [{"word": str(n), "hint": count_decimal_places(n)} for n in float_numbers]
put(
    "Floatle",
    float_numbers_mapped,
)
