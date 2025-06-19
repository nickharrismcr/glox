import time

def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n - 2) + fibonacci(n - 1)

total = 0
for j in range(10):
    start = time.time()
    for i in range(30):
        x = fibonacci(i)
    now = time.time() - start
    total += now
    print(j)
print(total / 10)