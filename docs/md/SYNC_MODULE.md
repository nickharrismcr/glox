# sync Module Documentation

The `sync` module provides `Mutex`, the one tool [`thread`](THREAD_MODULE.md) gives
you for the gap it deliberately leaves open: top-level `var`s, class statics, and
module attributes are shared across every thread, not isolated the way a spawned
closure's own captured locals are — `Mutex` is what makes touching one of those
safe from more than one thread at a time.

## Usage

```lox
import sync
```

## The basic pattern

```lox
import thread
import sync

counter = 0
lock = sync.Mutex()

func incrementer() {
    lock.locked(func() {
        counter = counter + 1
    })
}

threads = []
foreach (i in range(20)) {
    threads.append(thread.spawn(incrementer))
}
foreach (t in threads) { t.wait() }

print counter   // always exactly 20
```

Without the mutex, twenty threads all doing `counter = counter + 1` concurrently
would race — each read-modify-write isn't atomic, so some increments would be lost
depending on scheduling. Serialising every increment through one `Mutex` makes the
final count deterministic.

## Module functions

### `sync.Mutex()` → Mutex
Constructs a new, unlocked mutex.

## Mutex objects

A `Mutex` captured by a `thread.spawn()`-ed closure's upvalue is **shared, not
cloned**, across the new thread and the caller — unlike ordinary captured data,
which is deep-copied so each side gets its own isolated copy. A lock only means
something if every thread is contending for the *same* one.

### `m.acquire()` → nil
Blocks until the lock is free, then takes it. Pair it with `release()` in a
`finally` block to guarantee release on every path — normal completion, a
caught exception, or one that isn't caught here:

```lox
m.acquire()
try {
    // ... critical section ...
} finally {
    m.release()
}
```

Prefer `locked()` below unless you specifically need a lock that spans more
than one statement or function call.

### `m.release()` → nil
Releases the lock. Raises `SyncError` if called without a matching `acquire()`
(mirrors Go's own `sync.Mutex.Unlock()` panicking on the same misuse, caught and
turned into a catchable exception instead of crashing).

### `m.locked(closure)` → value
Convenience wrapper: acquires the lock, calls `closure` with no arguments, and
guarantees the lock is released when `closure` returns — whether normally or by
raising an uncaught exception — without needing an explicit `finally`. Returns
whatever `closure` returned.

If `closure` raises and the exception isn't caught *inside* `closure` itself, it
doesn't propagate with its original class the way a normal in-process exception
would — it surfaces as `SyncError` from `locked()` instead (the same "only the
message text survives the boundary" tradeoff `thread`'s `wait()`/`recv()` make, see
[THREAD_MODULE.md](THREAD_MODULE.md)). Catch it *inside* the closure with your own
`try`/`except` if you need the original exception's type preserved:

```lox
m.locked(func() {
    try {
        risky()
    } except SomeError as e {
        // handled here, with SomeError intact
    }
})
```

## Limitations

- v1 ships a single `Mutex` only — no `RWMutex`, semaphore, or `WaitGroup`-equivalent
  yet.
- `locked()`'s closure argument takes no parameters and its exception's specific
  class doesn't cross the boundary — see above.
