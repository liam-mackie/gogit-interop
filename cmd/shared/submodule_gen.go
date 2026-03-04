package main

/*
#include <stdlib.h>
*/
import "C"
import (
	git "github.com/go-git/go-git/v6"
)

//export GitSubmoduleInit
func GitSubmoduleInit(subHandle C.longlong) *C.char {
	sub, ok := loadHandle[*git.Submodule](int64(subHandle))
	if !ok {
		return C.CString("invalid submodule handle")
	}
	return toCError(sub.Init())
}

//export GitSubmoduleUpdate
func GitSubmoduleUpdate(subHandle C.longlong, optsHandle C.longlong) *C.char {
	sub, ok := loadHandle[*git.Submodule](int64(subHandle))
	if !ok {
		return C.CString("invalid submodule handle")
	}
	opts, ok := loadHandle[*git.SubmoduleUpdateOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid SubmoduleUpdateOptions handle")
	}
	return toCError(sub.Update(opts))
}

//export GitSubmoduleRepository
func GitSubmoduleRepository(subHandle C.longlong, handleOut *C.longlong) *C.char {
	sub, ok := loadHandle[*git.Submodule](int64(subHandle))
	if !ok {
		return C.CString("invalid submodule handle")
	}
	repo, err := sub.Repository()
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(repo))
	return nil
}

//export GitSubmoduleConfigName
func GitSubmoduleConfigName(subHandle C.longlong, nameOut **C.char) *C.char {
	sub, ok := loadHandle[*git.Submodule](int64(subHandle))
	if !ok {
		return C.CString("invalid submodule handle")
	}
	*nameOut = C.CString(sub.Config().Name)
	return nil
}

//export GitSubmoduleFree
func GitSubmoduleFree(subHandle C.longlong) {
	removeHandle(int64(subHandle))
}
