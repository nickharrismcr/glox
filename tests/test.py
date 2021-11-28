import sys, glob,subprocess,difflib
 

def run(fname):

    res = subprocess.Popen(["..\glox.exe","%s" % fname],stdout=subprocess.PIPE)
    return res

def basename(path):
 
    if "\\" in path:
        return path.split("\\")[-1]
    return path

def process(fname,write,verbose):

    pipe = run(fname)
    testdatafile="output/%s.testoutput" % basename(fname)
 
    if write:
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
                if verbose:
                    a=res[0].decode('ascii').splitlines()
                    b=data.decode('ascii').splitlines()
                    d=difflib.context_diff(a,b)
                    print ('\n'.join(d))

######################################################################################################################

write=False
verbose=False

if len(sys.argv) > 1 :
    if sys.argv[1] in ("--read","--write"):
        write=True if sys.argv[1]=="--write" else False
        del(sys.argv[1])
    if sys.argv[1] == "--verbose":
        verbose=True
        del(sys.argv[1])

if len(sys.argv) > 1 :
    process(sys.argv[1],write,verbose)
else:
    for f in glob.glob("lox/*lox"):
        process(f,write,verbose)
    