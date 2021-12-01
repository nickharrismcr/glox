# GLOX

**Bob Nystroms clox bytecode interpreter implemented in Go**

Cop out : GC is handled by the Go runtime.  

**Additions to vanilla Lox:**

immutable vars e.g  const a = 1;

integer number type:

      modulus operator %  

loop break/continue

native funcs :  

      int(number)     - conversion
      float(number)   - conversion 
      str(value)      - conversion 
      len(string|list) -> int
      sin(float)    -> float
      cos(float)    -> float 
      args() - returns list of command line arguments - not ideal! 

lists :

      initialiser (a=[]; a=[1,2,3];)
      indexing ( b=a[int] )
      index assign ( a[int] = b )
      slicing ( b=a[x:y]; b=a[:y]; b=a[x:]; b=a[:] )
      slice assignment ( e.g a[2:5] = [1,2,3] )
      adding ( list3=list1+list2 )
      append (list,val) (native)
      join ( list, string )    e.g join(["a","b","c"],"|") -> "a|b|c"  (native)

dictionaries: 
      initialiser ( a = {}; a = { "b":"c","d":"e"}; )
      get ( a[key])
      set ( a[key]=b)

strings :
      slices   ( a = "abcd"; b=a[0], b=a[:2], etc)
      multiply by integer ( a la python, e.g  "@" * 3 ,  3 * "@" = "@@@" )

renamed fun to func!

**TODO:**

# EASY 

Bob's classes chapter

list item del  (del a[b] or del a[b:c] - i.e assign nil )
  
  - sugar for a[b] = nil ? etc? 

allow class toString() magic method to define str()/print output

 - can't work while str() is implemented as a built-in as built in calls are expected to return a result to be pushed on to the stack
 - but str(class) needs to call an instance method in lox i.e new frame, run bytecode 
 - maybe make it a keyword/opcode like print? 

dictionary get keys as list 

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