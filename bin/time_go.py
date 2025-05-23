
import subprocess
import time
import sys

def time_go_program(go_binary_path, *args):
    start = time.perf_counter()
    
    try:
        result = subprocess.run(
            [go_binary_path, *args],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            check=True
        )
        end = time.perf_counter()
        duration = end - start

        print("Output:")
        print(result.stdout)
        print(f"\nTime: {duration:.4f} seconds")
    
    except subprocess.CalledProcessError as e:
        print("Error running Go program:")
        print(e.stderr)
        sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python time_go.py ./your_go_binary [args...]")
        sys.exit(1)
    
    time_go_program(sys.argv[1], *sys.argv[2:])
