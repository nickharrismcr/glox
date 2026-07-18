import pytest
from lox_helper import run_lox

BASIC_EXPECTED = [
    "false",
    "42",
    "0",
    "caught broken pipe",
    "done",
    "nil",
]

POOL_EXPECTED = [
    "4",
    "9",
    "25",
    "49",
    "121",
    "169",
    "289",
    "361",
    "done",
    "nil",
]


@pytest.mark.parametrize("force_compile", [False, True])
def test_process_basic(force_compile):
    lines = run_lox("process_basic.lox", force_compile=force_compile)
    assert lines == BASIC_EXPECTED


@pytest.mark.parametrize("force_compile", [False, True])
def test_process_pool(force_compile):
    lines = run_lox("process_pool.lox", force_compile=force_compile)
    assert lines == POOL_EXPECTED
