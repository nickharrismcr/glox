import subprocess
import time
import sys
import os

def time_lox_script(args):
    times = []
    for j in range(10):
        start = time.perf_counter()
        exe = os.path.join("bin", "glox.exe") if os.name == "nt" else os.path.join("bin", "glox")
        cmd = [exe] + args
        try:
            result = subprocess.run(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                text=True,
                check=True
            )
            end = time.perf_counter()
            duration = end - start
            times.append(duration)
            print(f"Run {j+1}: {duration:.4f} seconds")
        except subprocess.CalledProcessError as e:
            print("Error running Go program:")
            print(e.stdout)
            sys.exit(1)
    avg = sum(times) / len(times)
    print(f"Average: {avg:.4f} seconds")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python time_lox.py <script> [args...]")
        sys.exit(1)
    time_lox_script(sys.argv[1:])