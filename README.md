# GLOX

**Bob Nystroms CLox bytecode interpreter implemented in Go**

The aim of this project is to learn more deeply about programming in Go and the crafting of interpreters by way of implementing Bobs CLox interpreter in Go, adding Python-inspired extensions to Lox along the way.
The extensions to the language include enhanced string operations, lists, dictionaries, exception handling, module imports with bytecode caching, string and list iteration, and I/O.  

My implementation is very slow compared to CLox. The VM 
- does a lot of function calls in place of C macros, not all of which get inlined
- has a large switch/case inner loop which Go compiler doesn't optimise well  
- uses slow map[string] for globals - function code runs much quicker 
- uses interface{} for objects ( values are tagged union structs for speed but contain a pointer for objects ) 
- GC is handled by the Go runtime. 

but hey-ho. This is a learning exercise, the Go code is probably not very ideomatic. The fun is in figuring out how to get the interpreter to do new language stuff. 

**Additions to vanilla Lox:**

module imports

- e.g `import othermodule;`
- importing modules will cache the compiled bytecode in `__loxcache__/<module>.lxc`, subsequent imports will load from this cache unless the source is newer in which 
  case the module will be recompiled. 

EOL semicolons are optional 

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
- `type(var)` -> string - return type of variable e.g int,float,string,list,dict,class,instance,closure,file etc 
- `draw_png(filename,float array)` - generate a png using passed float array ( values 0 (black) to 1 (white)) 
- `encode_rgb(r,g,b) -> float `  - encode rgb int values (0-255) as a single float 
- `decode_rgb(f)`  - decode an encoded rgb float into a tuple of ints (0-255)     

native objects :

- fast 2D native float array 
- - `var a = float_array(100,100);`
- - `a.set(10,10,0.5);`
- - `var b=a.get(10,10);`

built-in lox modules:
-  iterator tools, function tools, string utils, PNG plotter 

lists :

- initialiser (`var a=[]; var a=[1,2,3];`)
- `l.append(val)` -> append val to list in place  
- indexing ( `b=a[int]` )
- index assign ( `a[int] = b` )
- slicing (`b=a[x:y]; b=a[:y]; b=a[x:]; b=a[:];` )
- slice assignment ( e.g `a[2:5] = [1,2,3];` )
- adding ( `list3=list1+list2;` )
- test for `item in list`  -> true|false 

tuples : 

- immutable lists
- `var a = (1,2,3);` 
- allows same operations as list but no append or assignment allowed.

dictionaries:

- initialiser ( `var a = {}; var a = { "b":"c","d":"e"};` )
- get ( `a[key]` or `a.get(key,default)` ) 
- set ( `a[key]=b`)
- `dict.keys()`   get list of keys 

strings :

- `s=s+"4"`  string addition
- `s.replace(old,new)` -> string  - replace substring in string   
- `s.join(list)` -> string - join each list item with s  e.g `join(["a","b","c"],"|");` -> "a|b|c"   
- multiply by integer ( a la python, e.g  `"@" * 3`,  `3 * "@"` = `"@@@"` )
- slices   ( `a = "abcd"; b=a[0], b=a[:2]`, etc )
- test for presence of substring in string : `substring in string` -> true|false 

foreach : 
- iterate lists/strings with `foreach ( var i in iterable ) { block }`



renamed fun to func

class `toString()` magic method

- if present and returns a string, will be used for print class / str(class)

exceptions

- built in Exception class, subclass custom exception classes from it
- `try {block} except [exception type] as [instance var] {handler}` 
- can nest try...excepts 
- can specify multiple handlers for different exception types
- `raise [exception instance]` statement 
- runtime can raise catchable exceptions e.g RunTimeError, EOFError

i/o

- native file open, close, readln, write 
- readln throws EOFError on eof 

**TODO:**

- more runtime exception types 
- exception stack trace 
- lxcs to store pointers to original source for error reporting 
- from module import [*|name] 
- import module as <namespace> 

- etc.
 
