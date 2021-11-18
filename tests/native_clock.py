from datetime import datetime

def fib(n):
  if (n < 2):
    return n;
  return fib(n - 2) + fib(n - 1)


start = datetime.now()
print (fib(35))
print ((datetime.now()-start).total_seconds())