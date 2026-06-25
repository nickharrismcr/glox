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

def time_lox_script(args):
    exe = os.path.join("bin", "glox.exe") if os.name == "nt" else os.path.join("bin", "glox")
    return time_command([exe] + args)

def time_python_script(args):
    return time_command([sys.executable] + args)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python time_lox.py [--python] <script> [args...]")
        sys.exit(1)
    if sys.argv[1] == "--python":
        time_python_script(sys.argv[2:])
    else:
        time_lox_script(sys.argv[1:])
