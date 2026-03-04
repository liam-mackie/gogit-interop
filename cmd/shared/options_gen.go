package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport"
)

//export GitCloneOptionsNew
func GitCloneOptionsNew(handleOut *C.longlong) {
	opts := &git.CloneOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitCloneOptionsSetURL
func GitCloneOptionsSetURL(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	opts.URL = C.GoString(val)
	return nil
}

//export GitCloneOptionsSetAuth
func GitCloneOptionsSetAuth(handle C.longlong, authHandle C.longlong) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	auth, ok := loadHandle[transport.AuthMethod](int64(authHandle))
	if !ok {
		return C.CString("invalid auth handle")
	}
	opts.Auth = auth
	return nil
}

//export GitCloneOptionsSetRemoteName
func GitCloneOptionsSetRemoteName(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	opts.RemoteName = C.GoString(val)
	return nil
}

//export GitCloneOptionsSetReferenceName
func GitCloneOptionsSetReferenceName(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	opts.ReferenceName = plumbing.ReferenceName(C.GoString(val))
	return nil
}

//export GitCloneOptionsSetSingleBranch
func GitCloneOptionsSetSingleBranch(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	opts.SingleBranch = val != 0
	return nil
}

//export GitCloneOptionsSetMirror
func GitCloneOptionsSetMirror(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	opts.Mirror = val != 0
	return nil
}

//export GitCloneOptionsSetNoCheckout
func GitCloneOptionsSetNoCheckout(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	opts.NoCheckout = val != 0
	return nil
}

//export GitCloneOptionsSetDepth
func GitCloneOptionsSetDepth(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	opts.Depth = int(val)
	return nil
}

//export GitCloneOptionsSetBare
func GitCloneOptionsSetBare(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	opts.Bare = val != 0
	return nil
}

//export GitCloneOptionsSetInsecureSkipTLS
func GitCloneOptionsSetInsecureSkipTLS(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	opts.InsecureSkipTLS = val != 0
	return nil
}

//export GitCloneOptionsSetShared
func GitCloneOptionsSetShared(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	opts.Shared = val != 0
	return nil
}

//export GitCloneOptionsSetTags
func GitCloneOptionsSetTags(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(handle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	opts.Tags = plumbing.TagMode(val)
	return nil
}

//export GitCloneOptionsFree
func GitCloneOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitPullOptionsNew
func GitPullOptionsNew(handleOut *C.longlong) {
	opts := &git.PullOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitPullOptionsSetRemoteName
func GitPullOptionsSetRemoteName(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.PullOptions](int64(handle))
	if !ok {
		return C.CString("invalid PullOptions handle")
	}
	opts.RemoteName = C.GoString(val)
	return nil
}

//export GitPullOptionsSetRemoteURL
func GitPullOptionsSetRemoteURL(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.PullOptions](int64(handle))
	if !ok {
		return C.CString("invalid PullOptions handle")
	}
	opts.RemoteURL = C.GoString(val)
	return nil
}

//export GitPullOptionsSetReferenceName
func GitPullOptionsSetReferenceName(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.PullOptions](int64(handle))
	if !ok {
		return C.CString("invalid PullOptions handle")
	}
	opts.ReferenceName = plumbing.ReferenceName(C.GoString(val))
	return nil
}

//export GitPullOptionsSetSingleBranch
func GitPullOptionsSetSingleBranch(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.PullOptions](int64(handle))
	if !ok {
		return C.CString("invalid PullOptions handle")
	}
	opts.SingleBranch = val != 0
	return nil
}

//export GitPullOptionsSetDepth
func GitPullOptionsSetDepth(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.PullOptions](int64(handle))
	if !ok {
		return C.CString("invalid PullOptions handle")
	}
	opts.Depth = int(val)
	return nil
}

//export GitPullOptionsSetAuth
func GitPullOptionsSetAuth(handle C.longlong, authHandle C.longlong) *C.char {
	opts, ok := loadHandle[*git.PullOptions](int64(handle))
	if !ok {
		return C.CString("invalid PullOptions handle")
	}
	auth, ok := loadHandle[transport.AuthMethod](int64(authHandle))
	if !ok {
		return C.CString("invalid auth handle")
	}
	opts.Auth = auth
	return nil
}

//export GitPullOptionsSetForce
func GitPullOptionsSetForce(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.PullOptions](int64(handle))
	if !ok {
		return C.CString("invalid PullOptions handle")
	}
	opts.Force = val != 0
	return nil
}

//export GitPullOptionsSetInsecureSkipTLS
func GitPullOptionsSetInsecureSkipTLS(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.PullOptions](int64(handle))
	if !ok {
		return C.CString("invalid PullOptions handle")
	}
	opts.InsecureSkipTLS = val != 0
	return nil
}

//export GitPullOptionsFree
func GitPullOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitFetchOptionsNew
func GitFetchOptionsNew(handleOut *C.longlong) {
	opts := &git.FetchOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitFetchOptionsSetRemoteName
func GitFetchOptionsSetRemoteName(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.FetchOptions](int64(handle))
	if !ok {
		return C.CString("invalid FetchOptions handle")
	}
	opts.RemoteName = C.GoString(val)
	return nil
}

//export GitFetchOptionsSetRemoteURL
func GitFetchOptionsSetRemoteURL(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.FetchOptions](int64(handle))
	if !ok {
		return C.CString("invalid FetchOptions handle")
	}
	opts.RemoteURL = C.GoString(val)
	return nil
}

//export GitFetchOptionsSetDepth
func GitFetchOptionsSetDepth(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.FetchOptions](int64(handle))
	if !ok {
		return C.CString("invalid FetchOptions handle")
	}
	opts.Depth = int(val)
	return nil
}

//export GitFetchOptionsSetAuth
func GitFetchOptionsSetAuth(handle C.longlong, authHandle C.longlong) *C.char {
	opts, ok := loadHandle[*git.FetchOptions](int64(handle))
	if !ok {
		return C.CString("invalid FetchOptions handle")
	}
	auth, ok := loadHandle[transport.AuthMethod](int64(authHandle))
	if !ok {
		return C.CString("invalid auth handle")
	}
	opts.Auth = auth
	return nil
}

//export GitFetchOptionsSetTags
func GitFetchOptionsSetTags(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.FetchOptions](int64(handle))
	if !ok {
		return C.CString("invalid FetchOptions handle")
	}
	opts.Tags = plumbing.TagMode(val)
	return nil
}

//export GitFetchOptionsSetForce
func GitFetchOptionsSetForce(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.FetchOptions](int64(handle))
	if !ok {
		return C.CString("invalid FetchOptions handle")
	}
	opts.Force = val != 0
	return nil
}

//export GitFetchOptionsSetInsecureSkipTLS
func GitFetchOptionsSetInsecureSkipTLS(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.FetchOptions](int64(handle))
	if !ok {
		return C.CString("invalid FetchOptions handle")
	}
	opts.InsecureSkipTLS = val != 0
	return nil
}

//export GitFetchOptionsSetPrune
func GitFetchOptionsSetPrune(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.FetchOptions](int64(handle))
	if !ok {
		return C.CString("invalid FetchOptions handle")
	}
	opts.Prune = val != 0
	return nil
}

//export GitFetchOptionsFree
func GitFetchOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitPushOptionsNew
func GitPushOptionsNew(handleOut *C.longlong) {
	opts := &git.PushOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitPushOptionsSetRemoteName
func GitPushOptionsSetRemoteName(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.PushOptions](int64(handle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	opts.RemoteName = C.GoString(val)
	return nil
}

//export GitPushOptionsSetRemoteURL
func GitPushOptionsSetRemoteURL(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.PushOptions](int64(handle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	opts.RemoteURL = C.GoString(val)
	return nil
}

//export GitPushOptionsSetAuth
func GitPushOptionsSetAuth(handle C.longlong, authHandle C.longlong) *C.char {
	opts, ok := loadHandle[*git.PushOptions](int64(handle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	auth, ok := loadHandle[transport.AuthMethod](int64(authHandle))
	if !ok {
		return C.CString("invalid auth handle")
	}
	opts.Auth = auth
	return nil
}

//export GitPushOptionsSetPrune
func GitPushOptionsSetPrune(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.PushOptions](int64(handle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	opts.Prune = val != 0
	return nil
}

//export GitPushOptionsSetForce
func GitPushOptionsSetForce(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.PushOptions](int64(handle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	opts.Force = val != 0
	return nil
}

//export GitPushOptionsSetInsecureSkipTLS
func GitPushOptionsSetInsecureSkipTLS(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.PushOptions](int64(handle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	opts.InsecureSkipTLS = val != 0
	return nil
}

//export GitPushOptionsSetFollowTags
func GitPushOptionsSetFollowTags(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.PushOptions](int64(handle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	opts.FollowTags = val != 0
	return nil
}

//export GitPushOptionsSetAtomic
func GitPushOptionsSetAtomic(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.PushOptions](int64(handle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	opts.Atomic = val != 0
	return nil
}

//export GitPushOptionsSetQuiet
func GitPushOptionsSetQuiet(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.PushOptions](int64(handle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	opts.Quiet = val != 0
	return nil
}

//export GitPushOptionsFree
func GitPushOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitCheckoutOptionsNew
func GitCheckoutOptionsNew(handleOut *C.longlong) {
	opts := &git.CheckoutOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitCheckoutOptionsSetHash
func GitCheckoutOptionsSetHash(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.CheckoutOptions](int64(handle))
	if !ok {
		return C.CString("invalid CheckoutOptions handle")
	}
	h := plumbing.NewHash(C.GoString(val))
	opts.Hash = h
	return nil
}

//export GitCheckoutOptionsSetBranch
func GitCheckoutOptionsSetBranch(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.CheckoutOptions](int64(handle))
	if !ok {
		return C.CString("invalid CheckoutOptions handle")
	}
	opts.Branch = plumbing.ReferenceName(C.GoString(val))
	return nil
}

//export GitCheckoutOptionsSetCreate
func GitCheckoutOptionsSetCreate(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CheckoutOptions](int64(handle))
	if !ok {
		return C.CString("invalid CheckoutOptions handle")
	}
	opts.Create = val != 0
	return nil
}

//export GitCheckoutOptionsSetForce
func GitCheckoutOptionsSetForce(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CheckoutOptions](int64(handle))
	if !ok {
		return C.CString("invalid CheckoutOptions handle")
	}
	opts.Force = val != 0
	return nil
}

//export GitCheckoutOptionsSetKeep
func GitCheckoutOptionsSetKeep(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CheckoutOptions](int64(handle))
	if !ok {
		return C.CString("invalid CheckoutOptions handle")
	}
	opts.Keep = val != 0
	return nil
}

//export GitCheckoutOptionsFree
func GitCheckoutOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitResetOptionsNew
func GitResetOptionsNew(handleOut *C.longlong) {
	opts := &git.ResetOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitResetOptionsSetCommit
func GitResetOptionsSetCommit(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.ResetOptions](int64(handle))
	if !ok {
		return C.CString("invalid ResetOptions handle")
	}
	h := plumbing.NewHash(C.GoString(val))
	opts.Commit = h
	return nil
}

//export GitResetOptionsSetMode
func GitResetOptionsSetMode(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.ResetOptions](int64(handle))
	if !ok {
		return C.CString("invalid ResetOptions handle")
	}
	opts.Mode = git.ResetMode(val)
	return nil
}

//export GitResetOptionsFree
func GitResetOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitRestoreOptionsNew
func GitRestoreOptionsNew(handleOut *C.longlong) {
	opts := &git.RestoreOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitRestoreOptionsSetStaged
func GitRestoreOptionsSetStaged(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.RestoreOptions](int64(handle))
	if !ok {
		return C.CString("invalid RestoreOptions handle")
	}
	opts.Staged = val != 0
	return nil
}

//export GitRestoreOptionsSetWorktree
func GitRestoreOptionsSetWorktree(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.RestoreOptions](int64(handle))
	if !ok {
		return C.CString("invalid RestoreOptions handle")
	}
	opts.Worktree = val != 0
	return nil
}

//export GitRestoreOptionsAddFile
func GitRestoreOptionsAddFile(handle C.longlong, path *C.char) *C.char {
	opts, ok := loadHandle[*git.RestoreOptions](int64(handle))
	if !ok {
		return C.CString("invalid RestoreOptions handle")
	}
	opts.Files = append(opts.Files, C.GoString(path))
	return nil
}

//export GitRestoreOptionsFree
func GitRestoreOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitCommitOptionsNew
func GitCommitOptionsNew(handleOut *C.longlong) {
	opts := &git.CommitOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitCommitOptionsSetAll
func GitCommitOptionsSetAll(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CommitOptions](int64(handle))
	if !ok {
		return C.CString("invalid CommitOptions handle")
	}
	opts.All = val != 0
	return nil
}

//export GitCommitOptionsSetAllowEmptyCommits
func GitCommitOptionsSetAllowEmptyCommits(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CommitOptions](int64(handle))
	if !ok {
		return C.CString("invalid CommitOptions handle")
	}
	opts.AllowEmptyCommits = val != 0
	return nil
}

//export GitCommitOptionsSetSigner
func GitCommitOptionsSetSigner(handle C.longlong, signerHandle C.longlong) *C.char {
	opts, ok := loadHandle[*git.CommitOptions](int64(handle))
	if !ok {
		return C.CString("invalid CommitOptions handle")
	}
	signer, ok := loadHandle[git.Signer](int64(signerHandle))
	if !ok {
		return C.CString("invalid signer handle")
	}
	opts.Signer = signer
	return nil
}

//export GitCommitOptionsSetAmend
func GitCommitOptionsSetAmend(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CommitOptions](int64(handle))
	if !ok {
		return C.CString("invalid CommitOptions handle")
	}
	opts.Amend = val != 0
	return nil
}

//export GitCommitOptionsSetAuthorNameEmail
func GitCommitOptionsSetAuthorNameEmail(handle C.longlong, name *C.char, email *C.char) *C.char {
	opts, ok := loadHandle[*git.CommitOptions](int64(handle))
	if !ok {
		return C.CString("invalid options handle")
	}
	opts.Author = &object.Signature{
		Name:  C.GoString(name),
		Email: C.GoString(email),
		When:  time.Now(),
	}
	return nil
}

//export GitCommitOptionsSetCommitterNameEmail
func GitCommitOptionsSetCommitterNameEmail(handle C.longlong, name *C.char, email *C.char) *C.char {
	opts, ok := loadHandle[*git.CommitOptions](int64(handle))
	if !ok {
		return C.CString("invalid options handle")
	}
	opts.Committer = &object.Signature{
		Name:  C.GoString(name),
		Email: C.GoString(email),
		When:  time.Now(),
	}
	return nil
}

//export GitCommitOptionsFree
func GitCommitOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitCreateTagOptionsNew
func GitCreateTagOptionsNew(handleOut *C.longlong) {
	opts := &git.CreateTagOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitCreateTagOptionsSetMessage
func GitCreateTagOptionsSetMessage(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.CreateTagOptions](int64(handle))
	if !ok {
		return C.CString("invalid CreateTagOptions handle")
	}
	opts.Message = C.GoString(val)
	return nil
}

//export GitCreateTagOptionsSetTaggerNameEmail
func GitCreateTagOptionsSetTaggerNameEmail(handle C.longlong, name *C.char, email *C.char) *C.char {
	opts, ok := loadHandle[*git.CreateTagOptions](int64(handle))
	if !ok {
		return C.CString("invalid options handle")
	}
	opts.Tagger = &object.Signature{
		Name:  C.GoString(name),
		Email: C.GoString(email),
		When:  time.Now(),
	}
	return nil
}

//export GitCreateTagOptionsFree
func GitCreateTagOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitAddOptionsNew
func GitAddOptionsNew(handleOut *C.longlong) {
	opts := &git.AddOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitAddOptionsSetAll
func GitAddOptionsSetAll(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.AddOptions](int64(handle))
	if !ok {
		return C.CString("invalid AddOptions handle")
	}
	opts.All = val != 0
	return nil
}

//export GitAddOptionsSetPath
func GitAddOptionsSetPath(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.AddOptions](int64(handle))
	if !ok {
		return C.CString("invalid AddOptions handle")
	}
	opts.Path = C.GoString(val)
	return nil
}

//export GitAddOptionsSetGlob
func GitAddOptionsSetGlob(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.AddOptions](int64(handle))
	if !ok {
		return C.CString("invalid AddOptions handle")
	}
	opts.Glob = C.GoString(val)
	return nil
}

//export GitAddOptionsSetSkipStatus
func GitAddOptionsSetSkipStatus(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.AddOptions](int64(handle))
	if !ok {
		return C.CString("invalid AddOptions handle")
	}
	opts.SkipStatus = val != 0
	return nil
}

//export GitAddOptionsFree
func GitAddOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitCleanOptionsNew
func GitCleanOptionsNew(handleOut *C.longlong) {
	opts := &git.CleanOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitCleanOptionsSetDir
func GitCleanOptionsSetDir(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.CleanOptions](int64(handle))
	if !ok {
		return C.CString("invalid CleanOptions handle")
	}
	opts.Dir = val != 0
	return nil
}

//export GitCleanOptionsFree
func GitCleanOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitLogOptionsNew
func GitLogOptionsNew(handleOut *C.longlong) {
	opts := &git.LogOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitLogOptionsSetFrom
func GitLogOptionsSetFrom(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.LogOptions](int64(handle))
	if !ok {
		return C.CString("invalid LogOptions handle")
	}
	h := plumbing.NewHash(C.GoString(val))
	opts.From = h
	return nil
}

//export GitLogOptionsSetTo
func GitLogOptionsSetTo(handle C.longlong, val *C.char) *C.char {
	opts, ok := loadHandle[*git.LogOptions](int64(handle))
	if !ok {
		return C.CString("invalid LogOptions handle")
	}
	h := plumbing.NewHash(C.GoString(val))
	opts.To = h
	return nil
}

//export GitLogOptionsSetOrder
func GitLogOptionsSetOrder(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.LogOptions](int64(handle))
	if !ok {
		return C.CString("invalid LogOptions handle")
	}
	opts.Order = git.LogOrder(val)
	return nil
}

//export GitLogOptionsSetAll
func GitLogOptionsSetAll(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.LogOptions](int64(handle))
	if !ok {
		return C.CString("invalid LogOptions handle")
	}
	opts.All = val != 0
	return nil
}

//export GitLogOptionsFree
func GitLogOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitPlainOpenOptionsNew
func GitPlainOpenOptionsNew(handleOut *C.longlong) {
	opts := &git.PlainOpenOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitPlainOpenOptionsSetDetectDotGit
func GitPlainOpenOptionsSetDetectDotGit(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.PlainOpenOptions](int64(handle))
	if !ok {
		return C.CString("invalid PlainOpenOptions handle")
	}
	opts.DetectDotGit = val != 0
	return nil
}

//export GitPlainOpenOptionsFree
func GitPlainOpenOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitListOptionsNew
func GitListOptionsNew(handleOut *C.longlong) {
	opts := &git.ListOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitListOptionsSetAuth
func GitListOptionsSetAuth(handle C.longlong, authHandle C.longlong) *C.char {
	opts, ok := loadHandle[*git.ListOptions](int64(handle))
	if !ok {
		return C.CString("invalid ListOptions handle")
	}
	auth, ok := loadHandle[transport.AuthMethod](int64(authHandle))
	if !ok {
		return C.CString("invalid auth handle")
	}
	opts.Auth = auth
	return nil
}

//export GitListOptionsSetInsecureSkipTLS
func GitListOptionsSetInsecureSkipTLS(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.ListOptions](int64(handle))
	if !ok {
		return C.CString("invalid ListOptions handle")
	}
	opts.InsecureSkipTLS = val != 0
	return nil
}

//export GitListOptionsSetTimeout
func GitListOptionsSetTimeout(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.ListOptions](int64(handle))
	if !ok {
		return C.CString("invalid ListOptions handle")
	}
	opts.Timeout = int(val)
	return nil
}

//export GitListOptionsFree
func GitListOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitMergeOptionsNew
func GitMergeOptionsNew(handleOut *C.longlong) {
	opts := &git.MergeOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitMergeOptionsSetStrategy
func GitMergeOptionsSetStrategy(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.MergeOptions](int64(handle))
	if !ok {
		return C.CString("invalid MergeOptions handle")
	}
	opts.Strategy = git.MergeStrategy(val)
	return nil
}

//export GitMergeOptionsFree
func GitMergeOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitSubmoduleUpdateOptionsNew
func GitSubmoduleUpdateOptionsNew(handleOut *C.longlong) {
	opts := &git.SubmoduleUpdateOptions{}
	*handleOut = C.longlong(storeHandle(opts))
}

//export GitSubmoduleUpdateOptionsSetInit
func GitSubmoduleUpdateOptionsSetInit(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.SubmoduleUpdateOptions](int64(handle))
	if !ok {
		return C.CString("invalid SubmoduleUpdateOptions handle")
	}
	opts.Init = val != 0
	return nil
}

//export GitSubmoduleUpdateOptionsSetNoFetch
func GitSubmoduleUpdateOptionsSetNoFetch(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.SubmoduleUpdateOptions](int64(handle))
	if !ok {
		return C.CString("invalid SubmoduleUpdateOptions handle")
	}
	opts.NoFetch = val != 0
	return nil
}

//export GitSubmoduleUpdateOptionsSetAuth
func GitSubmoduleUpdateOptionsSetAuth(handle C.longlong, authHandle C.longlong) *C.char {
	opts, ok := loadHandle[*git.SubmoduleUpdateOptions](int64(handle))
	if !ok {
		return C.CString("invalid SubmoduleUpdateOptions handle")
	}
	auth, ok := loadHandle[transport.AuthMethod](int64(authHandle))
	if !ok {
		return C.CString("invalid auth handle")
	}
	opts.Auth = auth
	return nil
}

//export GitSubmoduleUpdateOptionsSetDepth
func GitSubmoduleUpdateOptionsSetDepth(handle C.longlong, val C.int) *C.char {
	opts, ok := loadHandle[*git.SubmoduleUpdateOptions](int64(handle))
	if !ok {
		return C.CString("invalid SubmoduleUpdateOptions handle")
	}
	opts.Depth = int(val)
	return nil
}

//export GitSubmoduleUpdateOptionsFree
func GitSubmoduleUpdateOptionsFree(handle C.longlong) {
	removeHandle(int64(handle))
}

var (
	_ = time.Now
	_ object.Signature
)
