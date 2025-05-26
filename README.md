# GLOX

**Bob Nystroms clox bytecode interpreter implemented in Go**

Is slow compared to CLox. VM 
- does a lot of function calls in place of C macros, not all of which get inlined
- has a large switch/case inner loop which Go compiler doesn't optimise well and isn't cache friendly 
- uses slow map[string] for globals - function code runs much quicker 
- uses interface{} for objects ( values are tagged union structs for speed but contain a pointer for objects ) 
- GC is handled by the Go runtime. 

but hey-ho. 

**Additions to vanilla Lox:**

module imports

- e.g `import othermodule;`

immutable vars e.g  `const a = 1;`

integer number type:

- modulus operator %  

loop `break`/`continue`

native funcs :  

- `int(number)`    - conversion
- `float(number)`   - conversion 
- `str(value)`     - conversion 
- `len(string|list)` -> int
- `sin(float)`    -> float
- `cos(float)`    -> float 
- `args()` - returns list of command line arguments  
- `replace(string,old,new)` -> string  - replace substring of string  

lists :

- initialiser (`a=[]; a=[1,2,3];`)
- `append (list,val)`  
- indexing ( `b=a[int]` )
- index assign ( `a[int] = b` )
- slicing (`b=a[x:y]; b=a[:y]; b=a[x:]; b=a[:];` )
- slice assignment ( e.g `a[2:5] = [1,2,3];` )
- adding ( `list3=list1+list2;` )
- `join ( list, string )`    e.g `join(["a","b","c"],"|");` -> "a|b|c"   

dictionaries:

- initialiser ( `a = {}; a = { "b":"c","d":"e"};` )
- get ( `a[key]`)
- set ( `a[key]=b`)
- `keys(dict)`   get list of keys 

strings :

- multiply by integer ( a la python, e.g  `"@" * 3`,  `3 * "@"` = `"@@@"` )
- slices   ( `a = "abcd"; b=a[0], b=a[:2]`, etc )

renamed fun to func

class `toString()` magic method

- if present and returns a string, will be used for print class / str(class)

exceptions

- built in Exception class, subclass custom exception classes from it
- `try {block} except [exception type] as [instance var] {handler}` 
- can nest try...excepts 
- can specify multiple handlers for different exception types
- `raise [exception instance]` statement 
- runtime can raise catchable exceptions e.g EOFError

i/o

- native file open, close, readln, write 
- readln throws EOFError on eof 

**TODO:**

- more runtime exception types 
- marshal / unmarshal code chunks to/from .lxc files 
- from module import [*|name] 
- import module as <namespace> 
- foreach <iterator>
- etc.
 