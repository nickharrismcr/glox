# GLOX

**Bob Nystroms clox bytecode interpreter implemented in Go**

Cop out : GC is handled by the Go runtime.  

**Additions to vanilla Lox:**

immutable vars e.g  const a = 1;

integer number type 

modulus operator %   ( integers only )

loop break/continue

string multiply by integer ( a la python, e.g  "@" * 3 ,  3 * "@" = "@@@" )

native funcs :  

      int(number)     - conversion
      float(number)   - conversion 
      str(value)      - conversion 
      len(string|list) -> int
      sin(float)    -> float
      cos(float)    -> float 
      append(list,value)    
      args() - returns list of command line arguments - not ideal! 

lists :

      initialiser (a=[]; a=[1,2,3];)
      indexing ( b=a[int] )
      index assign ( a[int] = b )
      slicing ( b=a[x:y]; b=a[:y]; b=a[x:]; b=a[:] )
      slice assignment ( e.g a[2:5] = [1,2,3] )
      adding ( list3=list1+list2 )
      appends ( native append(list,val) )
      join ( list, string )    e.g join(["a","b","c"],"|") -> "a|b|c"

string slices   ( a = "abcd"; b=a[0], b=a[:2], etc)

renamed fun to func!

**TODO:**

# EASY 

Bob's classes chapter

list item del  (del a[b] or del a[b:c] - i.e assign nil )
  
  - sugar for a[b] = nil ? etc? 



allow class __str__ magic method to define str()/print output

dictionaries (maps)

# MEDIUM

module imports / dot for imported module access ( e.g sys.argv ) 
 
marshal / unmarshal code chunks to/from .lxc files 

# HARD

map/filter function
- can't do this as with native funcs as passed function is a closure object and needs to run bytecode 
- need to be language functions with new opcodes ? or vm function that manipulates stack 
- python map takes variable number of args : 1 = function,  2+ = iterables, function arity must equal no of iterable args 

- GENERAL : how to implement built ins that can take/run closure args ?

list comprehensions 
- build on map function ^^^

exceptions 
- try / catch 

etc etc 