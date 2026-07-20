import pytest
from lox_helper import run_lox

REUSE_EXPECTED = [
    "[ 4 , 9 , 16 ]",
    "[ 25 , 36 , 49 , 64 , 81 ]",
    "true",
    "caught closed pool",
    "done",
    "nil",
]


@pytest.mark.parametrize("force_compile", [False, True])
def test_pool_reuse(force_compile):
    # A Pool must be reusable across multiple map() calls without
    # respawning its workers (proven by matching pid()s before/after two
    # differently-sized map() calls), and map() on a closed pool must raise
    # PoolError rather than hang or crash.
    lines = run_lox("pool_reuse.lox", force_compile=force_compile)
    assert lines == REUSE_EXPECTED
