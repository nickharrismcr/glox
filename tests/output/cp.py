import sys,os,glob

for f in glob.glob("*output"):
    a,b,c = f.split(".")
    new=f"{a}_ns.{b}.{c}"
    print(new)
    os.system(f"cp {f} {new}")
