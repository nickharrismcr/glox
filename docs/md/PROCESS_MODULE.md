# process Module Documentation

The `process` module spawns other `glox` processes and communicates with them over a
pipe carrying [pickled](PICKLE_MODULE.md) values — glox's answer to Python's
`multiprocessing`. Since glox has no `fork()` equivalent and no way to serialise a
closure, each worker is a separate `.lox` script run as its own OS process, not an
in-memory function.

## Usage

```lox
import process
```

## The basic pattern

```lox
// parent.lox
import process

p = process.spawn("worker.lox")
p.send(42)
print p.recv()      // result from the worker
p.wait()
```

```lox
// worker.lox
import process

p = process.parent()
x = p.recv()
p.send(x * 2)
```

## Module functions

### `process.spawn(script_path, ...)` → Process
Launches another `glox` process running `script_path`, connected to the caller by a
pipe on the child's stdin/stdout. Any extra arguments become the child's own
`sys.args()` (as if it had been invoked directly from the command line).
- **script_path**: path to the `.lox` script the child should run
- Extra arguments must be strings, and **must not start with `-`** — `main.go`'s
  command-line flag parser has no `--` escape hatch, so a leading `-` would be
  misread as a flag and abort the child.

### `process.parent()` → Process
Returns a `Process` wired to this process's own stdin/stdout — the far end of the
pipe the process that spawned this one is holding. Used inside a worker script to
talk back to whatever spawned it. Only exposes `send`/`recv`/`try_recv` (there's no
underlying child to `wait()`/`kill()`/query the `pid()` of).

### `process.wait_any(processes)` → (index, value) or nil
Blocks until *any* of the given `Process` objects has a message ready, and returns
which one (as an index into `processes`) plus the received value. This is how you
fan-in results from several workers without polling each one in turn — the same
role Python's `multiprocessing.connection.wait()` plays.

If one of the processes has simply finished (its script ran to completion and
closed its end of the pipe cleanly), that's not treated as an error for the wait
as a whole — it's dropped from consideration and `wait_any` keeps waiting on
whichever of the others are still live. Once *every* process in the list has
finished, `wait_any` returns `nil` rather than raising — a live result is always
the 2-tuple `(index, value)`, so `nil` is an unambiguous "the whole pool is done"
signal for the caller to check, not an exceptional condition. A genuine I/O
problem (a broken pipe, a truncated message) still raises `ProcessError`
immediately.

## Process objects

### `p.send(value)` → nil
Serialises `value` ([pickle](PICKLE_MODULE.md) rules apply — plain data only, no
closures/instances/native objects) and writes it to the pipe.

### `p.recv()` → value
Blocks until a value arrives on the pipe and returns it. Raises `ProcessError` if
the peer has closed its end (e.g. the other process exited) or the pipe errors.

### `p.try_recv()` → (ok, value)
Non-blocking version of `recv()`. Returns `(false, nil)` immediately if nothing is
waiting, or `(true, value)` if a message was ready — the same shape as Go's own
`v, ok := <-channel` idiom.

### `p.wait()` → int
Blocks until the child process exits and returns its exit code. Only available on
a `spawn()`-side `Process` (not `parent()`).

### `p.kill()` → nil
Forcibly terminates the child process. Only available on a `spawn()`-side `Process`.

### `p.pid()` → int
The child process's OS process ID. Only available on a `spawn()`-side `Process`.

## Worker pool: draining a shared task queue

There's no shared-memory `Queue` the way Python's `multiprocessing.Queue` needs one
— but `spawn`/`send`/`wait_any` are enough to build the same worker-pool pattern,
and the `pool` module (see the language reference's `pool` section) does exactly
that, so you don't have to hand-roll it:

```lox
import pool

p = pool.Pool("worker.lox", 3)               // spawns 3 workers up front
results = p.map([2, 3, 5, 7, 11, 13, 17, 19]) // more tasks than workers
p.close()
```

`pool.Pool` is reusable — call `.map()` again on the same pool (workers stay
alive between calls) rather than paying spawn cost per batch.

Under the hood, `Pool.map()` is just this pattern: the "queue" is an ordinary
list living in the parent (only one process can hold it); the parent hands out
the next task the moment `wait_any` reports a worker is free. Worth knowing if
you need something `pool` doesn't offer (e.g. tasks that arrive incrementally
rather than as one batch):

```lox
import process

tasks = [2, 3, 5, 7, 11, 13, 17, 19]        // more tasks than workers
workers = []
foreach (i in range(3)) { workers.append(process.spawn("worker.lox")) }

next_task = 0
foreach (w in workers) {
    w.send(tasks[next_task])
    next_task = next_task + 1
}

outstanding = len(workers)
while (outstanding > 0) {
    idx, result = process.wait_any(workers)
    print result
    if (next_task < len(tasks)) {
        workers[idx].send(tasks[next_task])
        next_task = next_task + 1
    } else {
        outstanding = outstanding - 1
    }
}
foreach (w in workers) { w.wait() }
```

Each `wait_any` return *is* the "this worker is idle, give it more work" signal —
there's no separate "ready for work" message needed from the worker side.

## Limitations

- No closures: a worker is always a `.lox` script path, never an in-memory function
  (see [pickle](PICKLE_MODULE.md) for why).
- Blocking I/O only: `recv()` blocks the whole interpreter until data arrives —
  glox has no concurrency/async model to do otherwise. `wait_any` is the one
  multiplexing primitive available.
- No auto-reap: if a script never calls `wait()`/`kill()` on a spawned process, glox
  does nothing special about it, the same as Python expects `Process.join()`.
- Extra `spawn()` arguments can't start with `-` (see above).
