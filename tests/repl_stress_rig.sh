#!/usr/bin/env bash
# REPL stress rig for glox.
#
# Feeds a mixture of correct code and every category of error (syntax,
# undefined var/method/module, type errors, arithmetic edge cases, control-flow
# misuse, recursion) through `glox --repl`. Each snippet runs in its OWN --repl
# process so a Go panic in one cannot hide the others.
#
# Classification per snippet:
#   CRASH - a Go-level panic (goroutine/panic:/runtime error:/nil pointer) => real bug
#   ok    - survived to the @@ALIVE sentinel (Lox runtime/compile errors are expected)
#   DIED  - no crash text but never reached @@ALIVE (REPL wedged / exited early)
#
# Usage:  bash tests/repl_stress_rig.sh
#
# Note on operators: glox is strongly typed — `+`/`-`/`*`/`/` are arithmetic
# only, and string concatenation uses `&`. Section A below uses the correct forms.

set -u
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export LOX_PATH="$REPO_ROOT"
GLOX="$REPO_ROOT/bin/glox"
cd "$REPO_ROOT" || exit 1

crashes=0
run() {
  local name="$1"; local code="$2"
  local out
  out=$(printf '%s\nprint "@@ALIVE";\n\n' "$code" | "$GLOX" --repl 2>&1)
  if echo "$out" | grep -Eq 'goroutine |panic:|runtime error:|invalid memory address|nil pointer|divide by zero|index out of range'; then
    echo "=== CRASH: $name ==="
    echo "$out" | grep -Ev '^GLOX:|^> |^\.\.\. ' | head -8
    echo
    crashes=$((crashes+1))
  elif ! echo "$out" | grep -q '@@ALIVE'; then
    echo "=== DIED (no sentinel): $name ==="
    echo "$out" | tail -4
    echo
  fi
}

########################  A. CORRECT CODE  ########################
run A01_arith        'print 1 + 2 * 3 - 4 / 2;'
run A02_intdiv       'print 10 / 3;'
run A03_mod          'print 17 % 5;'
run A04_floatmix     'print 2.5 + 1; print 3 * 2.0;'
run A05_strcat       'var s = "abc"; print s & "def";'
run A06_interp       'var n = "world"; print "hi ${n} ${1+2}";'
run A07_list         'var xs = [1,2,3]; xs.append(4); print xs; print xs[-1];'
run A08_slice        'var xs = [1,2,3,4,5]; print xs[1:4]; print xs[:2]; print xs[3:];'
run A09_dict         'var d = {"a":1,"b":2}; d["c"]=3; print d["b"]; print len(d);'
run A10_tuple        'var t = (1,2,3); print t; print t[0];'
run A11_if           'var x = 5; if (x > 3) { print "big"; } else { print "small"; }'
run A12_while        'var i = 0; while (i < 3) { print i; i = i + 1; }'
run A13_for          'for (var i = 0; i < 3; i = i + 1) { print i; }'
run A14_foreach      'foreach (var v in [10,20,30]) { print v; }'
run A15_range        'foreach (var i in range(3)) { print i; }'
run A16_breakcont    'foreach (var i in range(5)) { if (i == 2) { continue; } if (i == 4) { break; } print i; }'
run A17_func         'func add(a,b) { return a + b; } print add(3,4);'
run A18_closure      'func counter() { var c = 0; return func() { c = c + 1; return c; }; } var f = counter(); print f(); print f();'
run A19_recursion    'func fib(n) { if (n < 2) { return n; } return fib(n-1) + fib(n-2); } print fib(10);'
run A20_lambda       'var sq = func(x) { return x*x; }; print sq(9);'
run A21_higherorder  'func apply(f, x) { return f(x); } print apply(func(n){ return n+1; }, 41);'
run A22_class        'class C { init(v) { this.v = v; } get() { return this.v; } } var c = C(42); print c.get();'
run A23_inherit      'class Animal { speak() { return "..."; } } class Dog < Animal { speak() { return "woof"; } } print Dog().speak();'
run A24_super        'class A { greet() { return "A"; } } class B < A { greet() { return super.greet() & "B"; } } print B().greet();'
run A25_const        'const PI = 3.14; print PI;'
run A26_destructure  'a, b, c = [1,2,3]; print a + b + c;'
run A27_ternary      'var x = 5; print x > 3 ? "yes" : "no";'
run A28_compound     'var x = 10; x += 5; x *= 2; print x;'
run A29_fieldclosure 'class R { init(fn) { this.fn = fn; } go(x) { return this.fn(x); } } print R(func(n){ return n*10; }).go(5);'
run A30_toString     'class P { init(x) { this.x = x; } toString() { return "P(" & str(this.x) & ")"; } } print P(7);'
run A31_module       'import math; print math.floor(3.7);'

########################  B. RUNTIME ERRORS (should raise, not crash)  ########################
run B01_undefvar     'print nonexistent;'
run B02_undefvar_fn  'func f() { return missing; } f();'
run B03_undefmethod  'class C {} C().nope();'
run B04_undeffield   'class C {} print C().nofield;'
run B05_callnoncall  'var x = 5; x();'
run B06_callnumfield 'var n = 3; n.foo();'
run B07_undefmodule  'import nosuchmod;'
run B08_fromundef    'from nosuchmod import x;'
run B09_wrongargs    'func f(a,b) { return a+b; } f(1);'
run B10_wrongargs2   'func f(a) { return a; } f(1,2,3);'
run B11_typeadd      'print "a" + 5;'
run B12_typeaddlist  'print [1] + 5;'
run B13_indexoob     'var xs = [1,2,3]; print xs[10];'
run B14_indexneg     'var xs = [1,2,3]; print xs[-10];'
run B15_dictmiss     'var d = {"a":1}; print d["z"];'
run B16_notiterable  'foreach (var x in 5) { print x; }'
run B17_constreassign 'const K = 1; K = 2;'
run B18_nilfield     'var x = nil; print x.foo;'
run B19_nilcall      'var x = nil; x();'
run B20_intdivzero   'print 5 / 0;'
run B21_intmodzero   'print 5 % 0;'
run B22_floatdivzero 'print 5.0 / 0.0;'
run B23_strindexoob  'print "abc"[10];'
run B24_thistoplevel 'print this;'
run B25_supertoplevel 'class A { m() { return super.m(); } } A().m();'
run B26_breakoutside 'break;'
run B27_contoutside  'continue;'
run B28_returntop    'return 5;'
run B29_wronginit    'class C { init(a) { this.a = a; } } var c = C();'
run B30_concat_mismatch 'print 1 & "x";'

########################  C. SYNTAX / COMPILE ERRORS (complete lines) ########################
run C01_badexpr      'print 1 +;'
run C02_badassign    'var = 5;'
run C03_baddot       'print 5 . ;'
run C04_doublecomma  'print [1,,2];'
run C05_badkeyword   'class 123 {}'
run C06_emptyparen   'print ();'
run C07_reservedvar  'var if = 5;'
run C08_trailingop   'print 3 * ;'

########################  D. EDGE / STRESS  ########################
run D01_deeprecur    'func rec(n) { return rec(n+1); } rec(0);'
run D02_nestedstruct 'var d = {"a":[1,{"b":(2,3)}]}; print d["a"][1]["b"][1];'
run D03_emptylist    'print [];  print len([]);'
run D04_emptydict    'print {}; print len({});'
run D05_bignum       'print 999999999 * 999999999;'
run D06_chaincall    'class C { m() { return this; } } print C().m().m().m();'
run D07_selfref      'var xs = [1]; xs.append(xs); print len(xs);'
run D08_manyargs     'func f(a,b,c,d,e) { return a+b+c+d+e; } print f(1,2,3,4,5);'
run D09_shadowbuiltin 'var len = 5; print len;'
run D10_redeffunc    'func f() { return 1; } func f() { return 2; } print f();'
run D11_emptyclass   'class E {} print E();'
run D12_nestedclosure 'func a() { func b() { func c() { return 42; } return c(); } return b(); } print a();'

if [ "$crashes" -eq 0 ]; then
  echo "=== RIG COMPLETE: no crashes ==="
else
  echo "=== RIG COMPLETE: $crashes CRASH(es) ==="
fi
exit "$crashes"
