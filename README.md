# GLOX

**Bob Nystroms clox bytecode interpreter implemented in Go**

Cop out : GC is handled by the Go runtime.  

**Additions to vanilla Lox:**

module imports

- e.g import othermodule;

immutable vars e.g  const a = 1;

integer number type:

- modulus operator %  

loop break/continue

native funcs :  

- int(number)     - conversion
- float(number)   - conversion 
- str(value)      - conversion 
- len(string|list) -> int
- sin(float)    -> float
- cos(float)    -> float 
- args() - returns list of command line arguments - not ideal! 

lists :

- initialiser (a=[]; a=[1,2,3];)
- append (list,val)  
- indexing ( b=a[int] )
- index assign ( a[int] = b )
- slicing ( b=a[x:y]; b=a[:y]; b=a[x:]; b=a[:] )
- slice assignment ( e.g a[2:5] = [1,2,3] )
- adding ( list3=list1+list2 )
- join ( list, string )    e.g join(["a","b","c"],"|") -> "a|b|c"   

dictionaries:

- initialiser ( a = {}; a = { "b":"c","d":"e"}; )
- get ( a[key])
- set ( a[key]=b)
- keys(dict)   get list of keys 

strings :

- multiply by integer ( a la python, e.g  "@" * 3 ,  3 * "@" = "@@@" )
- slices   ( a = "abcd"; b=a[0], b=a[:2], etc )

renamed fun to func

class toString() magic method

- if present and returns a string, will be used for print class / str(class)

**TODO:**

# EASY 

Bob's optimisation chapter

list item del  (del a[b] or del a[b:c] - i.e assign nil )
  
  - sugar for a[b] = nil ? etc? 


# MEDIUM
 
marshal / unmarshal code chunks to/from .lxc files 

# HARD

map/filter function
- implement in lox and make available as library? - needs module import function
- move to native? 

list comprehensions 
- build on map function ^^^

exceptions 
- try / catch 

etc etc 