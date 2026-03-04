package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/transport"
)

//export GitPlainInit
func GitPlainInit(path *C.char, isBare C.int, handleOut *C.longlong) *C.char {
	repo, err := git.PlainInit(C.GoString(path), isBare != 0)
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(repo))
	return nil
}

//export GitPlainOpen
func GitPlainOpen(path *C.char, handleOut *C.longlong) *C.char {
	repo, err := git.PlainOpen(C.GoString(path))
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(repo))
	return nil
}

//export GitPlainOpenWithOptions
func GitPlainOpenWithOptions(path *C.char, optsHandle C.longlong, handleOut *C.longlong) *C.char {
	opts, ok := loadHandle[*git.PlainOpenOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid PlainOpenOptions handle")
	}
	repo, err := git.PlainOpenWithOptions(C.GoString(path), opts)
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(repo))
	return nil
}

//export GitPlainClone
func GitPlainClone(path *C.char, optsHandle C.longlong, handleOut *C.longlong) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	repo, err := git.PlainClone(C.GoString(path), opts)
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(repo))
	return nil
}

//export GitRepositoryHead
func GitRepositoryHead(repoHandle C.longlong, refNameOut **C.char, hashOut **C.char) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	ref, err := repo.Head()
	if err != nil {
		return toCError(err)
	}
	*refNameOut = C.CString(string(ref.Name()))
	*hashOut = C.CString(ref.Hash().String())
	return nil
}

//export GitRepositoryWorktree
func GitRepositoryWorktree(repoHandle C.longlong, wtHandleOut *C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	wt, err := repo.Worktree()
	if err != nil {
		return toCError(err)
	}
	*wtHandleOut = C.longlong(storeHandle(wt))
	return nil
}

//export GitRepositoryFetch
func GitRepositoryFetch(repoHandle C.longlong, optsHandle C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	opts, ok := loadHandle[*git.FetchOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid FetchOptions handle")
	}
	return toCError(repo.Fetch(opts))
}

//export GitRepositoryPush
func GitRepositoryPush(repoHandle C.longlong, optsHandle C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	opts, ok := loadHandle[*git.PushOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	return toCError(repo.Push(opts))
}

//export GitRepositoryLog
func GitRepositoryLog(repoHandle C.longlong, optsHandle C.longlong, iterOut *C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	var opts *git.LogOptions
	if int64(optsHandle) != 0 {
		var ok bool
		opts, ok = loadHandle[*git.LogOptions](int64(optsHandle))
		if !ok {
			return C.CString("invalid LogOptions handle")
		}
	} else {
		opts = &git.LogOptions{}
	}
	iter, err := repo.Log(opts)
	if err != nil {
		return toCError(err)
	}
	*iterOut = C.longlong(storeHandle(iter))
	return nil
}

//export GitRepositoryTags
func GitRepositoryTags(repoHandle C.longlong, iterOut *C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	iter, err := repo.Tags()
	if err != nil {
		return toCError(err)
	}
	*iterOut = C.longlong(storeHandle(iter))
	return nil
}

//export GitRepositoryBranches
func GitRepositoryBranches(repoHandle C.longlong, iterOut *C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	iter, err := repo.Branches()
	if err != nil {
		return toCError(err)
	}
	*iterOut = C.longlong(storeHandle(iter))
	return nil
}

//export GitRepositoryNotes
func GitRepositoryNotes(repoHandle C.longlong, iterOut *C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	iter, err := repo.Notes()
	if err != nil {
		return toCError(err)
	}
	*iterOut = C.longlong(storeHandle(iter))
	return nil
}

//export GitRepositoryReferences
func GitRepositoryReferences(repoHandle C.longlong, iterOut *C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	iter, err := repo.References()
	if err != nil {
		return toCError(err)
	}
	*iterOut = C.longlong(storeHandle(iter))
	return nil
}

//export GitRepositoryReference
func GitRepositoryReference(repoHandle C.longlong, name *C.char, resolved C.int, refNameOut **C.char, hashOut **C.char) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	ref, err := repo.Reference(plumbing.ReferenceName(C.GoString(name)), resolved != 0)
	if err != nil {
		return toCError(err)
	}
	*refNameOut = C.CString(string(ref.Name()))
	*hashOut = C.CString(ref.Hash().String())
	return nil
}

//export GitRepositoryResolveRevision
func GitRepositoryResolveRevision(repoHandle C.longlong, rev *C.char, hashOut **C.char) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	hash, err := repo.ResolveRevision(plumbing.Revision(C.GoString(rev)))
	if err != nil {
		return toCError(err)
	}
	*hashOut = C.CString(hash.String())
	return nil
}

//export GitRepositoryCreateRemote
func GitRepositoryCreateRemote(repoHandle C.longlong, name *C.char, url *C.char, handleOut *C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: C.GoString(name),
		URLs: []string{C.GoString(url)},
	})
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(remote))
	return nil
}

//export GitRepositoryRemote
func GitRepositoryRemote(repoHandle C.longlong, name *C.char, handleOut *C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	remote, err := repo.Remote(C.GoString(name))
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(remote))
	return nil
}

//export GitRepositoryDeleteRemote
func GitRepositoryDeleteRemote(repoHandle C.longlong, name *C.char) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	return toCError(repo.DeleteRemote(C.GoString(name)))
}

//export GitRepositoryCreateBranch
func GitRepositoryCreateBranch(repoHandle C.longlong, name *C.char, hash *C.char) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	branch := &config.Branch{
		Name: C.GoString(name),
	}
	return toCError(repo.CreateBranch(branch))
}

//export GitRepositoryDeleteBranch
func GitRepositoryDeleteBranch(repoHandle C.longlong, name *C.char) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	return toCError(repo.DeleteBranch(C.GoString(name)))
}

//export GitRepositoryCreateTag
func GitRepositoryCreateTag(repoHandle C.longlong, name *C.char, hash *C.char, optsHandle C.longlong, refNameOut **C.char, hashOut **C.char) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	h := plumbing.NewHash(C.GoString(hash))
	var opts *git.CreateTagOptions
	if int64(optsHandle) != 0 {
		var ok bool
		opts, ok = loadHandle[*git.CreateTagOptions](int64(optsHandle))
		if !ok {
			return C.CString("invalid CreateTagOptions handle")
		}
	}
	ref, err := repo.CreateTag(C.GoString(name), h, opts)
	if err != nil {
		return toCError(err)
	}
	*refNameOut = C.CString(string(ref.Name()))
	*hashOut = C.CString(ref.Hash().String())
	return nil
}

//export GitRepositoryTag
func GitRepositoryTag(repoHandle C.longlong, name *C.char, refNameOut **C.char, hashOut **C.char) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	ref, err := repo.Tag(C.GoString(name))
	if err != nil {
		return toCError(err)
	}
	*refNameOut = C.CString(string(ref.Name()))
	*hashOut = C.CString(ref.Hash().String())
	return nil
}

//export GitRepositoryDeleteTag
func GitRepositoryDeleteTag(repoHandle C.longlong, name *C.char) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	return toCError(repo.DeleteTag(C.GoString(name)))
}

//export GitRepositoryCommitObject
func GitRepositoryCommitObject(repoHandle C.longlong, hash *C.char, commitHashOut **C.char, msgOut **C.char, authorNameOut **C.char, authorEmailOut **C.char, tsOut *C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	h := plumbing.NewHash(C.GoString(hash))
	commit, err := repo.CommitObject(h)
	if err != nil {
		return toCError(err)
	}
	*commitHashOut = C.CString(commit.Hash.String())
	*msgOut = C.CString(commit.Message)
	*authorNameOut = C.CString(commit.Author.Name)
	*authorEmailOut = C.CString(commit.Author.Email)
	*tsOut = C.longlong(commit.Author.When.Unix())
	return nil
}

//export GitRepositoryMerge
func GitRepositoryMerge(repoHandle C.longlong, refName *C.char, hash *C.char, optsHandle C.longlong) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	h := plumbing.NewHash(C.GoString(hash))
	ref := plumbing.NewHashReference(plumbing.ReferenceName(C.GoString(refName)), h)
	var opts git.MergeOptions
	if int64(optsHandle) != 0 {
		optsPtr, ok := loadHandle[*git.MergeOptions](int64(optsHandle))
		if !ok {
			return C.CString("invalid MergeOptions handle")
		}
		opts = *optsPtr
	}
	return toCError(repo.Merge(*ref, opts))
}

//export GitRepositoryRemotes
func GitRepositoryRemotes(repoHandle C.longlong, jsonOut **C.char) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	remotes, err := repo.Remotes()
	if err != nil {
		return toCError(err)
	}
	names := make([]string, len(remotes))
	for i, r := range remotes {
		names[i] = r.Config().Name
	}
	data, err := json.Marshal(names)
	if err != nil {
		return toCError(err)
	}
	*jsonOut = C.CString(string(data))
	return nil
}

//export GitRepositoryFree
func GitRepositoryFree(repoHandle C.longlong) {
	removeHandle(int64(repoHandle))
}

var (
	_ = json.Marshal
	_ config.RemoteConfig
	_ transport.AuthMethod
)
