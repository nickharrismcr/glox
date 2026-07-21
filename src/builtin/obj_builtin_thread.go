package builtin

import "glox/src/core"

// ThreadObject is the parent-side handle returned by thread.spawn() --
// the "I spawned this thread" view, exposing send/recv/wait/cancel.
type ThreadObject struct {
	core.BuiltInObject
	Handle  *core.ThreadHandle
	Methods map[int]*core.BuiltInObject
}

func newThreadObject(handle *core.ThreadHandle) *ThreadObject {
	return &ThreadObject{Handle: handle}
}

func (o *ThreadObject) String() string {
	return "<thread>"
}

func (o *ThreadObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *ThreadObject) GetNativeType() core.NativeType {
	return core.NATIVE_THREAD
}

func (o *ThreadObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}

func (o *ThreadObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (o *ThreadObject) IsBuiltIn() bool {
	return true
}

// ThreadChannelObject is the worker-side handle returned by
// thread.channel(), called from inside a spawned function -- the "talk
// back to whoever spawned me" view, exposing send/recv/try_recv.
type ThreadChannelObject struct {
	core.BuiltInObject
	Chans   *core.ThreadChannels
	Methods map[int]*core.BuiltInObject
}

func newThreadChannelObject(chans *core.ThreadChannels) *ThreadChannelObject {
	return &ThreadChannelObject{Chans: chans}
}

func (o *ThreadChannelObject) String() string {
	return "<thread channel>"
}

func (o *ThreadChannelObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *ThreadChannelObject) GetNativeType() core.NativeType {
	return core.NATIVE_THREAD_CHANNEL
}

func (o *ThreadChannelObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}

func (o *ThreadChannelObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (o *ThreadChannelObject) IsBuiltIn() bool {
	return true
}
