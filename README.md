# GLOX

**Bob Nystroms clox bytecode interpreter implemented in Go**

Cop out : GC is handled by the Go runtime.  

**Additions to vanilla Lox:**

immutable vars e.g  const a = 1;

modulus operator %

loop break/continue

string multiply by number ( a la python, e.g  "@" * 3 ,  3 * "@" = "@@@" )

native funcs :  

      str(value)    
      len(string|list)      
      sin(x)    
      cos(x)     
      append(list,value)    
      args() - returns list of command line arguments - not ideal! 

lists :

      initialiser (a=[]; a=[1,2,3];)
      indexing ( b=a[x] )
      index assign ( a[x] = b )
      slicing ( b=a[x:y]; b=a[:y]; b=a[x:]; b=a[:] )
      slice assignment ( e.g a[2:5] = [1,2,3] )
      adding ( list3=list1+list2 )
      appends ( native append(list,val) )

string slices   ( a = "abcd"; b=a[0], b=a[:2], etc)

renamed fun to func!

**TODO:**

separate int type 

Bob's classes chapter

-  allow class __str__ magic method to define str()/print output

imports / dot for imported module access ( e.g sys.argv ) 
 
marshal / unmarshal bytecode chunks to/from file

list item del  (del a[b] or del a[b:c] - i.e assign nil )
  
  - sugar for a[b] = nil ? etc? 
