from lox_helper import run_lox


def test_plusequals():
    lines = run_lox("plusequals.lox")
    assert lines[0] == "2"   # a=1, a+=1 → 2
    assert lines[1] == "3"   # b=2, b+=1 → 3
    assert lines[2] == "4"   # c=vec2(3,2), c.x+=1 → 4
    assert lines[-1] == "nil"
