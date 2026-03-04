package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

//export GitWorktreeStatus
func GitWorktreeStatus(wtHandle C.longlong, jsonOut **C.char) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	status, err := wt.Status()
	if err != nil {
		return toCError(err)
	}
	type fileStatusJSON struct {
		Staging  string `json:"staging"`
		Worktree string `json:"worktree"`
		Extra    string `json:"extra,omitempty"`
	}
	out := make(map[string]fileStatusJSON, len(status))
	for path, fs := range status {
		out[path] = fileStatusJSON{
			Staging:  string(fs.Staging),
			Worktree: string(fs.Worktree),
			Extra:    fs.Extra,
		}
	}
	data, err := json.Marshal(out)
	if err != nil {
		return toCError(err)
	}
	*jsonOut = C.CString(string(data))
	return nil
}

//export GitWorktreeAdd
func GitWorktreeAdd(wtHandle C.longlong, path *C.char, hashOut **C.char) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	hash, err := wt.Add(C.GoString(path))
	if err != nil {
		return toCError(err)
	}
	*hashOut = C.CString(hash.String())
	return nil
}

//export GitWorktreeAddWithOptions
func GitWorktreeAddWithOptions(wtHandle C.longlong, optsHandle C.longlong) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	opts, ok2 := loadHandle[*git.AddOptions](int64(optsHandle))
	if !ok2 {
		return C.CString("invalid AddOptions handle")
	}
	return toCError(wt.AddWithOptions(opts))
}

//export GitWorktreeAddGlob
func GitWorktreeAddGlob(wtHandle C.longlong, pattern *C.char) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	return toCError(wt.AddGlob(C.GoString(pattern)))
}

//export GitWorktreeCommit
func GitWorktreeCommit(wtHandle C.longlong, msg *C.char, optsHandle C.longlong, hashOut **C.char) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	var opts *git.CommitOptions
	if int64(optsHandle) != 0 {
		var ok2 bool
		opts, ok2 = loadHandle[*git.CommitOptions](int64(optsHandle))
		if !ok2 {
			return C.CString("invalid CommitOptions handle")
		}
	} else {
		opts = &git.CommitOptions{}
	}
	hash, err := wt.Commit(C.GoString(msg), opts)
	if err != nil {
		return toCError(err)
	}
	*hashOut = C.CString(hash.String())
	return nil
}

//export GitWorktreeCheckout
func GitWorktreeCheckout(wtHandle C.longlong, optsHandle C.longlong) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	opts, ok2 := loadHandle[*git.CheckoutOptions](int64(optsHandle))
	if !ok2 {
		return C.CString("invalid CheckoutOptions handle")
	}
	return toCError(wt.Checkout(opts))
}

//export GitWorktreePull
func GitWorktreePull(wtHandle C.longlong, optsHandle C.longlong) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	opts, ok2 := loadHandle[*git.PullOptions](int64(optsHandle))
	if !ok2 {
		return C.CString("invalid PullOptions handle")
	}
	return toCError(wt.Pull(opts))
}

//export GitWorktreeReset
func GitWorktreeReset(wtHandle C.longlong, optsHandle C.longlong) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	opts, ok2 := loadHandle[*git.ResetOptions](int64(optsHandle))
	if !ok2 {
		return C.CString("invalid ResetOptions handle")
	}
	return toCError(wt.Reset(opts))
}

//export GitWorktreeRestore
func GitWorktreeRestore(wtHandle C.longlong, optsHandle C.longlong) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	opts, ok2 := loadHandle[*git.RestoreOptions](int64(optsHandle))
	if !ok2 {
		return C.CString("invalid RestoreOptions handle")
	}
	return toCError(wt.Restore(opts))
}

//export GitWorktreeClean
func GitWorktreeClean(wtHandle C.longlong, optsHandle C.longlong) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	opts, ok2 := loadHandle[*git.CleanOptions](int64(optsHandle))
	if !ok2 {
		return C.CString("invalid CleanOptions handle")
	}
	return toCError(wt.Clean(opts))
}

//export GitWorktreeMove
func GitWorktreeMove(wtHandle C.longlong, fromPath *C.char, toPath *C.char, hashOut **C.char) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	hash, err := wt.Move(C.GoString(fromPath), C.GoString(toPath))
	if err != nil {
		return toCError(err)
	}
	*hashOut = C.CString(hash.String())
	return nil
}

//export GitWorktreeSubmodule
func GitWorktreeSubmodule(wtHandle C.longlong, name *C.char, handleOut *C.longlong) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	sub, err := wt.Submodule(C.GoString(name))
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(sub))
	return nil
}

//export GitWorktreeSubmodules
func GitWorktreeSubmodules(wtHandle C.longlong, jsonOut **C.char) *C.char {
	wt, ok := loadHandle[*git.Worktree](int64(wtHandle))
	if !ok {
		return C.CString("invalid worktree handle")
	}
	subs, err := wt.Submodules()
	if err != nil {
		return toCError(err)
	}
	names := make([]string, len(subs))
	for i, s := range subs {
		names[i] = s.Config().Name
	}
	data, err := json.Marshal(names)
	if err != nil {
		return toCError(err)
	}
	*jsonOut = C.CString(string(data))
	return nil
}

//export GitWorktreeFree
func GitWorktreeFree(wtHandle C.longlong) {
	removeHandle(int64(wtHandle))
}

var (
	_ = json.Marshal
	_ = time.Now
	_ object.Signature
	_ plumbing.Hash
)
