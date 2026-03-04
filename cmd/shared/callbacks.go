package main

/*
#include "callbacks.h"
*/
import "C"
import "unsafe"

type progressWriter struct {
	fn       C.GitProgressFunc
	userData unsafe.Pointer
}

func (w *progressWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	C.callProgressFunc(w.fn, (*C.char)(unsafe.Pointer(&p[0])), C.int(len(p)), w.userData)
	return len(p), nil
}
