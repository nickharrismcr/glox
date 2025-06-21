import sys, glob,subprocess,difflib,argparse,os
 

def run(fname,force_compile=False) :  

    cmdlst = ["D:/go/glox/bin/glox"]
    if force_compile:
        cmdlst.append("--force-compile")
    cmdlst.append(fname)
 
    res = subprocess.Popen(cmdlst,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
    return res

def basename(path):
 
    if "\\" in path:
        return path.split("\\")[-1]
    if "/" in path:
        return path.split("/")[-1]
    return path

def format(s):

    return "\n".join([ str(i.decode("ascii")) for i in s.splitlines() ])


def process(fname,args,force_compile=False,show_result=False):

    if args.verbose or show_result:
        print ("Test %-30s" % fname,end='')
        sys.stdout.flush()
        
    
    passed=False
    pipe = run(fname,force_compile)  
    testdatafile="output/%s.testoutput" % basename(fname)
 
    if args.write:
        with open(testdatafile,"wb") as outfile:
            res=pipe.communicate()
            outfile.write(res[0])
            passed=True
    else:
        with open(testdatafile,"rb") as infile:
            res=pipe.communicate()
            data=infile.read()
            match=data==res[0]
            if match:
                passed=True
                if args.verbose or show_result:
                    print ( ": PASS")
            else:
                if args.verbose or show_result:
                    print (": FAIL")
                    print (f'expecting:\n'+format(data))
                    print (f'got:\n'+format(res[0]))

                    a=res[0].decode('ascii').splitlines()
                    b=data.decode('ascii').splitlines()
        
                    if args.diff:
                        d=difflib.context_diff(a,b)
                        print ('\n'.join(d))
    
    return passed 

def process_all(test, args, force_compile=False):

    failed = []
    files = glob.glob(args.dir + "/*lox")
     
    for fname in files:
        if args.write: 
            testdatafile="output/%s.testoutput" % basename(fname)
            if not os.path.exists(testdatafile):
                print (f"Test {test} : {fname} : Creating test output file")
                process(fname, args, force_compile)
        elif not process(fname, args, force_compile):
            failed.append(fname)
    
    if args.write:
        return 
    
    if failed:
        print (f"{test} : One or more tests failed.")
        [ print(i) for i in failed ]
    else:
        print (f"{test} : All tests passed.")

######################################################################################################################

write=False
verbose=False

parser = argparse.ArgumentParser(description="Process .lox files with optional write and verbose modes.")
parser.add_argument("file", nargs="?", help="File to process (optional; if not provided, all lox/*lox files will be processed)")
parser.add_argument("--dir", nargs="?",default="lox")
parser.add_argument("--write", action="store_true", help="Enable write mode")
parser.add_argument("--verbose", action="store_true", help="Enable verbose output")
parser.add_argument("--diff", action="store_true", help="Show diff")

args = parser.parse_args()
failed=[]
if args.file:
    print(f"Test : {args.file} : Force import compile")
    process(args.file, args,show_result=True)
    print(f"Test : {args.file} : Import from lxc")
    process(args.file, args, show_result=True, force_compile=True)
else:
    process_all("Force import compile",args,force_compile=True)
    process_all("Import from lxc",args)
   

   