package main

/*
#include <stdlib.h>
*/
import "C"
import "unsafe"

func toCError(err error) *C.char {
	if err == nil {
		return nil
	}
	return C.CString(err.Error())
}

//export GitFreeString
func GitFreeString(s *C.char) {
	if s != nil {
		C.free(unsafe.Pointer(s))
	}
}
