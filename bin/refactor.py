import sys,os

mapping={
"FunctionObject":"obj_func", 
"ClosureObject":"obj_closure", 
"UpvalueObject":"obj_upval", 
"StringObject":"obj_string", 
"BuiltInObject":"obj_builtin", 
"ListObject":"obj_list", 
"DictObject":"obj_dict", 
"ClassObject":"obj_class", 
"InstanceObject":"obj_inst", 
"BoundMethodObject":"obj_method", 
"ModuleObject":"obj_module", 
"FileObject":"obj_file"
}

with open("lox/object.go") as inp:
    data=inp.read().splitlines()

def type_defn(row):

    if "type" in row and "struct" in row:
        name=row.split()[1]
        return True,name
    return False,""

def output(header,block):

    filename=mapping[block[0]]+".go"
    with open(filename,"w") as outp:
        for row in header:
            print (row,file=outp)
        for row in block[1:]:
            print (row,file=outp) 


objects={}
header=[]
top=True
inBlock=False 
block=[]
for row in data:
    
    ok,name =type_defn(row)
    print (ok,name)
    if ok and name in mapping:
        inBlock=True 
        if block != []:
            output(header,block)
        block=[name] 
    if top:
        header.append(row)
        if ")" in row:
            top=False
    elif inBlock:
        block.append(row)

