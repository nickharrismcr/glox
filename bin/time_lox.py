import subprocess
import time
import sys
import os

def time_command(cmd, runs=10):
    times = []
    for j in range(runs):
        start = time.perf_counter()
        try:
            subprocess.run(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                text=True,
                check=True
            )
            duration = time.perf_counter() - start
            times.append(duration)
            print(f"Run {j+1}: {duration:.4f} seconds")
        except subprocess.CalledProcessError as e:
            print("Error:", e.stdout)
            sys.exit(1)
    avg = sum(times) / len(times)
    print(f"Average: {avg:.4f} seconds")
    return avg

def time_lox_script(args, runs=10):
    exe = os.path.join("bin", "glox.exe") if os.name == "nt" else os.path.join("bin", "glox")
    return time_command([exe] + args, runs=runs)

def time_python_script(args, runs=10):
    return time_command([sys.executable] + args, runs=runs)

if __name__ == "__main__":
    args = sys.argv[1:]
    runs = 10
    if "--runs" in args:
        i = args.index("--runs")
        runs = int(args[i + 1])
        args = args[:i] + args[i + 2:]
    if not args:
        print("Usage: python time_lox.py [--python] [--runs N] <script> [args...]")
        sys.exit(1)
    if args[0] == "--python":
        time_python_script(args[1:], runs=runs)
    else:
        time_lox_script(args, runs=runs)
