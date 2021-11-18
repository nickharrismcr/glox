import sys, glob,subprocess,difflib
 

def run(fname):

    res = subprocess.Popen(["..\glox.exe","%s" % fname],stdout=subprocess.PIPE)
    return res

mode = "read"
if len(sys.argv) > 1 :
    mode = sys.argv[1]

for f in glob.glob("*lox"):

    pipe = run(f)
    testdatafile="output/%s.testoutput" % f
    if mode == "write":
        with open(testdatafile,"wb") as outfile:
            res=pipe.communicate()
            outfile.write(res[0])
    else:
        with open(testdatafile,"rb") as infile:
            res=pipe.communicate()
            data=infile.read()
            match=data==res[0]
            if match:
                print ("Test %-30s : PASS" % f)
            else:
                print ("Test %-30s : FAIL" % f)
                a=res[0].decode('ascii').splitlines()
                b=data.decode('ascii').splitlines()
                d=difflib.context_diff(a,b)
                print ('\n'.join(d))


            


