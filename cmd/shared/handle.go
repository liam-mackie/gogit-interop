package main

import (
	"sync"
	"sync/atomic"
)

var (
	handleCounter atomic.Int64
	handles       sync.Map
)

func storeHandle(obj any) int64 {
	h := handleCounter.Add(1)
	handles.Store(h, obj)
	return h
}

func loadHandle[T any](h int64) (T, bool) {
	v, ok := handles.Load(h)
	if !ok {
		var zero T
		return zero, false
	}
	t, ok := v.(T)
	return t, ok
}

func removeHandle(h int64) {
	handles.Delete(h)
}
