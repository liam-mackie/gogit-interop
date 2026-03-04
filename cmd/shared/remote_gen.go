package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"

	git "github.com/go-git/go-git/v6"
)

//export GitRemoteFetch
func GitRemoteFetch(remoteHandle C.longlong, optsHandle C.longlong) *C.char {
	remote, ok := loadHandle[*git.Remote](int64(remoteHandle))
	if !ok {
		return C.CString("invalid remote handle")
	}
	opts, ok := loadHandle[*git.FetchOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid FetchOptions handle")
	}
	return toCError(remote.Fetch(opts))
}

//export GitRemotePush
func GitRemotePush(remoteHandle C.longlong, optsHandle C.longlong) *C.char {
	remote, ok := loadHandle[*git.Remote](int64(remoteHandle))
	if !ok {
		return C.CString("invalid remote handle")
	}
	opts, ok := loadHandle[*git.PushOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	return toCError(remote.Push(opts))
}

//export GitRemoteList
func GitRemoteList(remoteHandle C.longlong, optsHandle C.longlong, jsonOut **C.char) *C.char {
	remote, ok := loadHandle[*git.Remote](int64(remoteHandle))
	if !ok {
		return C.CString("invalid remote handle")
	}
	opts, ok := loadHandle[*git.ListOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid ListOptions handle")
	}
	refs, err := remote.List(opts)
	if err != nil {
		return toCError(err)
	}
	type refJSON struct {
		Name string `json:"name"`
		Hash string `json:"hash"`
	}
	out := make([]refJSON, len(refs))
	for i, r := range refs {
		out[i] = refJSON{Name: string(r.Name()), Hash: r.Hash().String()}
	}
	data, err := json.Marshal(out)
	if err != nil {
		return toCError(err)
	}
	*jsonOut = C.CString(string(data))
	return nil
}

//export GitRemoteConfigName
func GitRemoteConfigName(remoteHandle C.longlong, nameOut **C.char) *C.char {
	remote, ok := loadHandle[*git.Remote](int64(remoteHandle))
	if !ok {
		return C.CString("invalid remote handle")
	}
	*nameOut = C.CString(remote.Config().Name)
	return nil
}

//export GitRemoteFree
func GitRemoteFree(remoteHandle C.longlong) {
	removeHandle(int64(remoteHandle))
}
