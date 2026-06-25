import pytest
from lox_helper import run_lox


@pytest.mark.parametrize("script", ["for_break.lox", "for_break_ns.lox"])
def test_for_break(script):
    lines = run_lox(script)
    assert lines[:-1] == ["1", "2", "3", "4", "5"]
    assert lines[-1] == "nil"


@pytest.mark.parametrize("script", ["for_break_nested.lox", "for_break_nested_ns.lox"])
def test_for_break_nested(script):
    lines = run_lox(script)
    # Outer loop breaks at a>5; inner loop breaks at aa>5
    # a=6 prints "6" then breaks (only one print before inner loop)
    assert "6" in lines
    assert "7" in lines
    assert lines[-1] == "nil"
    # a=1..5 each print a twice; a=6..9 print only once
    assert lines.count("1") == 2
    assert lines.count("6") == 1
