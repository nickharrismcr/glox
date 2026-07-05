from lox_helper import run_lox


EXPECTED = [
    "3",      # emit(3)
    "true",   # update() 1 — all alive (age 1 < life 2)
    "3",
    "false",  # update() 2 — all dead (age 2 >= life 2), swap-removed
    "0",
    "2",      # emit(2) — recycled from the pool
    "true",   # update() 1
    "false",  # update() 2
    "0",
]


def test_particle_sys_pool_and_removal():
    # Exercises particle_sys swap-remove removal and pool reuse (headless: the
    # module only imports random/math, no window needed).
    for force in (True, False):
        lines = run_lox("particle_sys.lox", force_compile=force)
        assert lines[-1] == "nil"      # top-level nil return
        assert lines[:-1] == EXPECTED
