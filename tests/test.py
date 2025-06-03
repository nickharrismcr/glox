import sys, glob,subprocess,difflib,argparse
 

def run(fname):

    res = subprocess.Popen(["../glox.exe","%s" % fname],stdout=subprocess.PIPE,stderr=subprocess.PIPE)
    return res

def basename(path):
 
    if "\\" in path:
        return path.split("\\")[-1]
    if "/" in path:
        return path.split("/")[-1]
    return path

def format(s):

    return "\n".join([ str(i.decode("ascii")) for i in s.splitlines() ])


def process(fname,args):

    
    pipe = run(fname)
    testdatafile="output/%s.testoutput" % basename(fname)
 
    if args.write:
        with open(testdatafile,"wb") as outfile:
            res=pipe.communicate()
            outfile.write(res[0])
    else:
        with open(testdatafile,"rb") as infile:
            res=pipe.communicate()
            data=infile.read()
            match=data==res[0]
            if match:
                print ("Test %-30s : PASS" % fname)
            else:
                print ("Test %-30s : FAIL" % fname)
            if args.verbose:
               
                print (f'expecting:\n'+format(data))
                print (f'got:\n'+format(res[0]))

                a=res[0].decode('ascii').splitlines()
                b=data.decode('ascii').splitlines()
    
                if args.diff:
                    d=difflib.context_diff(a,b)
                    print ('\n'.join(d))

######################################################################################################################

write=False
verbose=False

parser = argparse.ArgumentParser(description="Process .lox files with optional write and verbose modes.")
parser.add_argument("file", nargs="?", help="File to process (optional; if not provided, all lox/*lox files will be processed)")
parser.add_argument("--write", action="store_true", help="Enable write mode")
parser.add_argument("--verbose", action="store_true", help="Enable verbose output")
parser.add_argument("--diff", action="store_true", help="Show diff")

args = parser.parse_args()

if args.file:
    process(args.file, args)
else:
    for f in glob.glob("bin/*lox"):
        process(f, args)
