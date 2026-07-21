package core

// thread.go defines the plain-data types the thread module (see
// docs/thread-module-plan.md) uses to communicate between a spawning VM
// and a goroutine-backed worker VM. Kept in core (not vm) for the same
// reason ClassResolver/pickle live in core: src/builtin can't import
// src/vm (vm imports builtin to register built-ins -- a cycle), so
// anything a builtin needs to see has to be a core type reached through
// VMContext.

// ThreadMessage is what a spawned thread's own end (ThreadChannels.Out)
// sends back to the parent's ThreadHandle.FromWorker: either a value the
// worker chose to send, or (Err != nil) the terminal notification that the
// worker's goroutine ended abnormally (a recovered panic, or an unhandled
// Lox exception escaping run()).
type ThreadMessage struct {
	Val Value
	Err error
}

// ThreadChannels is the view of a spawn's communication channels handed to
// the *worker* side (see VMContext.ThreadChannels, called by
// thread.channel()). Field names are from the worker's point of view: In
// is what the worker reads, Out is what the worker writes.
type ThreadChannels struct {
	In        <-chan Value
	Out       chan<- ThreadMessage
	Cancelled <-chan struct{} // closed when the parent calls Thread.cancel()
}

// ThreadHandle is the view of a spawn's channels/lifecycle given to the
// *parent* (spawning) side (see VMContext.SpawnThread). Err and Result are
// only meaningful after Done is observed closed -- the worker goroutine
// writes them, then closes FromWorker, then Done, in that order, so Go's
// channel-close happens-before guarantee is the only synchronisation
// needed; no mutex required.
type ThreadHandle struct {
	ToWorker   chan<- Value
	FromWorker <-chan ThreadMessage
	Done       <-chan struct{}
	Cancel     func()
	Err        error
	Result     Value
}
