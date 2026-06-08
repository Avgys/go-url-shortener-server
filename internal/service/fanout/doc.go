// Package fanout delivers values from an input channel to subscriber channels.
//
// Producers send on [Fanout.In]. Subscribers receive on channels returned by [Fanout.Take].
// A background goroutine started by [New] reads from In until ctx is cancelled.
package fanout
