import sys, glob,subprocess,difflib
 

def run(fname):

    res = subprocess.Popen(["..\glox.exe","%s" % fname],stdout=subprocess.PIPE)
    return res

def process(fname,write):

    pipe = run(fname)
    testdatafile="output/%s.testoutput" % fname
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
                a=res[0].decode('ascii').splitlines()
                b=data.decode('ascii').splitlines()
                d=difflib.context_diff(a,b)
                print ('\n'.join(d))

######################################################################################################################

write=False

if len(sys.argv) > 1 :
    if sys.argv[1] in ("--read","--write"):
        write=True if sys.argv[1]=="--write" else False
        del(sys.argv[1])

if len(sys.argv) > 1 :
    process(sys.argv[1],write)
else:
    for f in glob.glob("*lox"):
        process(f,write)
    