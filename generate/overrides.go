package main

import (
	"fmt"
	"strings"
)

var overrideMethods = map[string]map[string]bool{
	"Repository": {
		"CreateRemote": true, "CreateBranch": true,
		"Merge": true, "Branch": true,
		"Config": true, "CreateRemoteAnonymous": true,
	},
	"Worktree": {
		"Status": true, "Submodules": true,
	},
	"Remote": {
		"List": true,
	},
	"Blob": {
		"Reader": true,
	},
	"Commit": {
		"Stats": true, "Patch": true,
		"MergeBase": true, "Verify": true,
	},
	"Tree": {
		"Diff": true, "Patch": true, "FindEntry": true,
	},
	"Tag": {
		"Verify": true,
	},
}

func isOverrideMethod(typeName, methodName string) bool {
	if methods, ok := overrideMethods[typeName]; ok {
		return methods[methodName]
	}
	return false
}

func generateOverrideMethod(b *strings.Builder, ht *HandleType, m Method) {
	switch ht.GoName + "." + m.GoName {
	case "Repository.CreateRemote":
		generateOverrideCreateRemote(b, m.CName)
	case "Repository.CreateBranch":
		generateOverrideCreateBranch(b, m.CName)
	case "Repository.Merge":
		generateOverrideMerge(b, m.CName)
	case "Repository.Branch":
		generateOverrideRepoBranch(b, m.CName)
	case "Repository.Config":
		generateOverrideRepoConfig(b, m.CName)
	case "Repository.CreateRemoteAnonymous":
		generateOverrideCreateRemoteAnonymous(b, m.CName)
	case "Worktree.Status":
		generateOverrideWorktreeStatus(b, m.CName)
	case "Worktree.Submodules":
		generateOverrideWorktreeSubmodules(b, m.CName)
	case "Remote.List":
		generateOverrideRemoteList(b, m.CName)
	case "Blob.Reader":
		generateOverrideBlobReader(b, m.CName)
	case "Commit.Stats":
		generateOverrideCommitStats(b, m.CName)
	case "Commit.Patch":
		generateOverrideCommitPatch(b, m.CName)
	case "Commit.MergeBase":
		generateOverrideCommitMergeBase(b, m.CName)
	case "Commit.Verify":
		generateOverrideCommitVerify(b, m.CName)
	case "Tree.Diff":
		generateOverrideTreeDiff(b, m.CName)
	case "Tree.Patch":
		generateOverrideTreePatch(b, m.CName)
	case "Tree.FindEntry":
		generateOverrideTreeFindEntry(b, m.CName)
	case "Tag.Verify":
		generateOverrideTagVerify(b, m.CName)
	}
}

func generateExtraMethodsGo(b *strings.Builder, ht *HandleType) {
	switch ht.GoName {
	case "Repository":
		generateExtraRepoRemotes(b)
		generateExtraCloneInMemory(b)
		generateExtraBlame(b)
	case "Remote":
		generateExtraRemoteConfigName(b)
		generateExtraRemoteConfig(b)
	case "Submodule":
		generateExtraSubmoduleConfigName(b)
		generateExtraSubmoduleConfig(b)
		generateExtraSubmoduleStatus(b)
	case "Commit":
		generateExtraCommitFieldAccessors(b)
	}
}

func writeOverrideNativeMethod(b *strings.Builder, ht *HandleType, m Method) {
	cName := m.CName
	switch ht.GoName + "." + m.GoName {
	case "Repository.CreateRemote":
		writeDllImport(b, cName, "long repoHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string name,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string url,\n        out long handleOut")
	case "Repository.CreateBranch":
		writeDllImport(b, cName, "long repoHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string name,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string hash")
	case "Repository.Merge":
		writeDllImport(b, cName, "long repoHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string refName,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string hash,\n        long optsHandle")
	case "Repository.Branch":
		writeDllImport(b, cName, "long repoHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string name,\n        out IntPtr jsonOut")
	case "Repository.Config":
		writeDllImport(b, cName, "long repoHandle, out IntPtr jsonOut")
	case "Repository.CreateRemoteAnonymous":
		writeDllImport(b, cName, "long repoHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string url,\n        out long handleOut")
	case "Worktree.Status":
		writeDllImport(b, cName, "long wtHandle, out IntPtr jsonOut")
	case "Worktree.Submodules":
		writeDllImport(b, cName, "long wtHandle, out IntPtr jsonOut")
	case "Remote.List":
		writeDllImport(b, cName, "long remoteHandle, long optsHandle, out IntPtr jsonOut")
	case "Blob.Reader":
		writeDllImport(b, cName, "long bHandle, out IntPtr dataOut")
	case "Commit.Stats":
		writeDllImport(b, cName, "long cHandle, out IntPtr jsonOut")
	case "Commit.Patch":
		writeDllImport(b, cName, "long cHandle, long toHandle, out IntPtr patchOut")
	case "Commit.MergeBase":
		writeDllImport(b, cName, "long cHandle, long otherHandle, out IntPtr jsonOut")
	case "Commit.Verify":
		writeDllImport(b, cName, "long cHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string armoredKeyRing")
	case "Tree.Diff":
		writeDllImport(b, cName, "long tHandle, long toHandle, out IntPtr jsonOut")
	case "Tree.Patch":
		writeDllImport(b, cName, "long tHandle, long toHandle, out IntPtr patchOut")
	case "Tree.FindEntry":
		writeDllImport(b, cName, "long tHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string path,\n        out IntPtr jsonOut")
	case "Tag.Verify":
		writeDllImport(b, cName, "long tHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string armoredKeyRing")
	default:
		writeGenericNativeMethod(b, ht, m)
	}
}

func writeExtraNativeMethods(b *strings.Builder, ht *HandleType) {
	switch ht.GoName {
	case "Repository":
		writeDllImport(b, "GitRepositoryRemotes", "long repoHandle, out IntPtr jsonOut")
		writeDllImport(b, "GitCloneInMemory", "long optsHandle, out long handleOut")
		writeDllImport(b, "GitBlame", "long commitHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string path,\n        out IntPtr jsonOut")
	case "Remote":
		writeDllImport(b, "GitRemoteConfigName", "long remoteHandle, out IntPtr nameOut")
		writeDllImport(b, "GitRemoteConfig", "long remoteHandle, out IntPtr jsonOut")
	case "Submodule":
		writeDllImport(b, "GitSubmoduleConfigName", "long subHandle, out IntPtr nameOut")
		writeDllImport(b, "GitSubmoduleConfig", "long subHandle, out IntPtr jsonOut")
		writeDllImport(b, "GitSubmoduleStatus", "long subHandle, out IntPtr jsonOut")
	case "Commit":
		for _, field := range []string{"Author", "Committer"} {
			for _, prop := range []string{"Name", "Email"} {
				writeDllImport(b, fmt.Sprintf("GitCommit%s%s", field, prop),
					fmt.Sprintf("long cHandle, out IntPtr %sOut", strings.ToLower(prop)))
			}
			writeDllImport(b, fmt.Sprintf("GitCommit%sWhen", field),
				"long cHandle, out long tsOut")
		}
	}
}

// --- Go override implementations ---

func loadReceiver(b *strings.Builder, typeName, handleParam string) {
	goType := fmt.Sprintf("*git.%s", typeName)
	fmt.Fprintf(b, "\trecv, ok := loadHandle[%s](int64(%s))\n", goType, handleParam)
	fmt.Fprintf(b, "\tif !ok {\n\t\treturn C.CString(\"invalid %s handle\")\n\t}\n", strings.ToLower(typeName))
}

func loadObjectReceiver(b *strings.Builder, typeName, handleParam string) {
	goType := fmt.Sprintf("*object.%s", typeName)
	fmt.Fprintf(b, "\trecv, ok := loadHandle[%s](int64(%s))\n", goType, handleParam)
	fmt.Fprintf(b, "\tif !ok {\n\t\treturn C.CString(\"invalid %s handle\")\n\t}\n", strings.ToLower(typeName))
}

func generateOverrideCreateRemote(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char, url *C.char, handleOut *C.longlong) *C.char {\n", cName)
	loadReceiver(b, "Repository", "repoHandle")
	b.WriteString("\tremote, err := recv.CreateRemote(&config.RemoteConfig{\n")
	b.WriteString("\t\tName: C.GoString(name),\n")
	b.WriteString("\t\tURLs: []string{C.GoString(url)},\n")
	b.WriteString("\t})\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*handleOut = C.longlong(storeHandle(remote))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideCreateBranch(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char, hash *C.char) *C.char {\n", cName)
	loadReceiver(b, "Repository", "repoHandle")
	b.WriteString("\tbranch := &config.Branch{\n")
	b.WriteString("\t\tName: C.GoString(name),\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn toCError(recv.CreateBranch(branch))\n}\n\n")
}

func generateOverrideMerge(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, refName *C.char, hash *C.char, optsHandle C.longlong) *C.char {\n", cName)
	loadReceiver(b, "Repository", "repoHandle")
	b.WriteString("\th := plumbing.NewHash(C.GoString(hash))\n")
	b.WriteString("\tref := plumbing.NewHashReference(plumbing.ReferenceName(C.GoString(refName)), h)\n")
	b.WriteString("\tvar opts git.MergeOptions\n")
	b.WriteString("\tif int64(optsHandle) != 0 {\n")
	b.WriteString("\t\toptsPtr, ok := loadHandle[*git.MergeOptions](int64(optsHandle))\n")
	b.WriteString("\t\tif !ok {\n\t\t\treturn C.CString(\"invalid MergeOptions handle\")\n\t\t}\n")
	b.WriteString("\t\topts = *optsPtr\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn toCError(recv.Merge(*ref, opts))\n}\n\n")
}

func generateOverrideWorktreeStatus(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, jsonOut **C.char) *C.char {\n", cName)
	loadReceiver(b, "Worktree", "wtHandle")
	b.WriteString("\tstatus, err := recv.Status()\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\ttype fileStatusJSON struct {\n")
	b.WriteString("\t\tStaging  string `json:\"staging\"`\n")
	b.WriteString("\t\tWorktree string `json:\"worktree\"`\n")
	b.WriteString("\t\tExtra    string `json:\"extra,omitempty\"`\n")
	b.WriteString("\t}\n")
	b.WriteString("\tout := make(map[string]fileStatusJSON, len(status))\n")
	b.WriteString("\tfor path, fs := range status {\n")
	b.WriteString("\t\tout[path] = fileStatusJSON{\n")
	b.WriteString("\t\t\tStaging:  string(fs.Staging),\n")
	b.WriteString("\t\t\tWorktree: string(fs.Worktree),\n")
	b.WriteString("\t\t\tExtra:    fs.Extra,\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("\tdata, err := json.Marshal(out)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*jsonOut = C.CString(string(data))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideWorktreeSubmodules(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, jsonOut **C.char) *C.char {\n", cName)
	loadReceiver(b, "Worktree", "wtHandle")
	b.WriteString("\tsubs, err := recv.Submodules()\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\tnames := make([]string, len(subs))\n")
	b.WriteString("\tfor i, s := range subs {\n")
	b.WriteString("\t\tnames[i] = s.Config().Name\n")
	b.WriteString("\t}\n")
	b.WriteString("\tdata, err := json.Marshal(names)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*jsonOut = C.CString(string(data))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideRemoteList(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(remoteHandle C.longlong, optsHandle C.longlong, jsonOut **C.char) *C.char {\n", cName)
	loadReceiver(b, "Remote", "remoteHandle")
	b.WriteString("\topts, ok := loadHandle[*git.ListOptions](int64(optsHandle))\n")
	b.WriteString("\tif !ok {\n\t\treturn C.CString(\"invalid ListOptions handle\")\n\t}\n")
	b.WriteString("\trefs, err := recv.List(opts)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\ttype refJSON struct {\n")
	b.WriteString("\t\tName string `json:\"name\"`\n")
	b.WriteString("\t\tHash string `json:\"hash\"`\n")
	b.WriteString("\t}\n")
	b.WriteString("\tout := make([]refJSON, len(refs))\n")
	b.WriteString("\tfor i, r := range refs {\n")
	b.WriteString("\t\tout[i] = refJSON{Name: string(r.Name()), Hash: r.Hash().String()}\n")
	b.WriteString("\t}\n")
	b.WriteString("\tdata, err := json.Marshal(out)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*jsonOut = C.CString(string(data))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideBlobReader(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(bHandle C.longlong, dataOut **C.char) *C.char {\n", cName)
	loadObjectReceiver(b, "Blob", "bHandle")
	b.WriteString("\treader, err := recv.Reader()\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\tdefer reader.Close()\n")
	b.WriteString("\tdata, err := io.ReadAll(reader)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*dataOut = C.CString(string(data))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

// --- Extra methods (not on Go types) ---

func generateExtraRepoRemotes(b *strings.Builder) {
	b.WriteString(`//export GitRepositoryRemotes
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

`)
}

func generateExtraRemoteConfigName(b *strings.Builder) {
	b.WriteString(`//export GitRemoteConfigName
func GitRemoteConfigName(remoteHandle C.longlong, nameOut **C.char) *C.char {
	remote, ok := loadHandle[*git.Remote](int64(remoteHandle))
	if !ok {
		return C.CString("invalid remote handle")
	}
	*nameOut = C.CString(remote.Config().Name)
	return nil
}

`)
}

func generateExtraSubmoduleConfigName(b *strings.Builder) {
	b.WriteString(`//export GitSubmoduleConfigName
func GitSubmoduleConfigName(subHandle C.longlong, nameOut **C.char) *C.char {
	sub, ok := loadHandle[*git.Submodule](int64(subHandle))
	if !ok {
		return C.CString("invalid submodule handle")
	}
	*nameOut = C.CString(sub.Config().Name)
	return nil
}

`)
}

func generateExtraCloneInMemory(b *strings.Builder) {
	b.WriteString(`//export GitCloneInMemory
func GitCloneInMemory(optsHandle C.longlong, handleOut *C.longlong) *C.char {
	opts, ok := loadHandle[*git.CloneOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid CloneOptions handle")
	}
	storer := memory.NewStorage()
	var wt billy.Filesystem
	if !opts.Bare {
		wt = memfs.New()
	}
	repo, err := git.Clone(storer, wt, opts)
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(repo))
	return nil
}

`)
}

// --- New override implementations ---

func generateOverrideRepoBranch(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char, jsonOut **C.char) *C.char {\n", cName)
	loadReceiver(b, "Repository", "repoHandle")
	b.WriteString("\tbranch, err := recv.Branch(C.GoString(name))\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\ttype branchJSON struct {\n")
	b.WriteString("\t\tName        string `json:\"name\"`\n")
	b.WriteString("\t\tRemote      string `json:\"remote\"`\n")
	b.WriteString("\t\tMerge       string `json:\"merge\"`\n")
	b.WriteString("\t\tRebase      string `json:\"rebase\"`\n")
	b.WriteString("\t\tDescription string `json:\"description,omitempty\"`\n")
	b.WriteString("\t}\n")
	b.WriteString("\tout := branchJSON{\n")
	b.WriteString("\t\tName:   branch.Name,\n")
	b.WriteString("\t\tRemote: branch.Remote,\n")
	b.WriteString("\t\tMerge:  string(branch.Merge),\n")
	b.WriteString("\t\tRebase: branch.Rebase,\n")
	b.WriteString("\t\tDescription: branch.Description,\n")
	b.WriteString("\t}\n")
	b.WriteString("\tdata, err := json.Marshal(out)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*jsonOut = C.CString(string(data))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideRepoConfig(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, jsonOut **C.char) *C.char {\n", cName)
	loadReceiver(b, "Repository", "repoHandle")
	b.WriteString("\tcfg, err := recv.Config()\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\ttype coreJSON struct {\n")
	b.WriteString("\t\tIsBare   bool   `json:\"isBare\"`\n")
	b.WriteString("\t\tWorktree string `json:\"worktree,omitempty\"`\n")
	b.WriteString("\t}\n")
	b.WriteString("\ttype identityJSON struct {\n")
	b.WriteString("\t\tName  string `json:\"name,omitempty\"`\n")
	b.WriteString("\t\tEmail string `json:\"email,omitempty\"`\n")
	b.WriteString("\t}\n")
	b.WriteString("\ttype initJSON struct {\n")
	b.WriteString("\t\tDefaultBranch string `json:\"defaultBranch,omitempty\"`\n")
	b.WriteString("\t}\n")
	b.WriteString("\ttype configJSON struct {\n")
	b.WriteString("\t\tCore      coreJSON     `json:\"core\"`\n")
	b.WriteString("\t\tUser      identityJSON `json:\"user\"`\n")
	b.WriteString("\t\tAuthor    identityJSON `json:\"author\"`\n")
	b.WriteString("\t\tCommitter identityJSON `json:\"committer\"`\n")
	b.WriteString("\t\tInit      initJSON     `json:\"init\"`\n")
	b.WriteString("\t}\n")
	b.WriteString("\tout := configJSON{\n")
	b.WriteString("\t\tCore: coreJSON{IsBare: cfg.Core.IsBare, Worktree: cfg.Core.Worktree},\n")
	b.WriteString("\t\tUser: identityJSON{Name: cfg.User.Name, Email: cfg.User.Email},\n")
	b.WriteString("\t\tAuthor: identityJSON{Name: cfg.Author.Name, Email: cfg.Author.Email},\n")
	b.WriteString("\t\tCommitter: identityJSON{Name: cfg.Committer.Name, Email: cfg.Committer.Email},\n")
	b.WriteString("\t\tInit: initJSON{DefaultBranch: string(cfg.Init.DefaultBranch)},\n")
	b.WriteString("\t}\n")
	b.WriteString("\tdata, err := json.Marshal(out)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*jsonOut = C.CString(string(data))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideCreateRemoteAnonymous(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, url *C.char, handleOut *C.longlong) *C.char {\n", cName)
	loadReceiver(b, "Repository", "repoHandle")
	b.WriteString("\tremote, err := recv.CreateRemoteAnonymous(&config.RemoteConfig{\n")
	b.WriteString("\t\tURLs: []string{C.GoString(url)},\n")
	b.WriteString("\t})\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*handleOut = C.longlong(storeHandle(remote))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideCommitStats(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(cHandle C.longlong, jsonOut **C.char) *C.char {\n", cName)
	loadObjectReceiver(b, "Commit", "cHandle")
	b.WriteString("\tstats, err := recv.Stats()\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\ttype fileStatJSON struct {\n")
	b.WriteString("\t\tName     string `json:\"name\"`\n")
	b.WriteString("\t\tAddition int    `json:\"addition\"`\n")
	b.WriteString("\t\tDeletion int    `json:\"deletion\"`\n")
	b.WriteString("\t}\n")
	b.WriteString("\tout := make([]fileStatJSON, len(stats))\n")
	b.WriteString("\tfor i, s := range stats {\n")
	b.WriteString("\t\tout[i] = fileStatJSON{Name: s.Name, Addition: s.Addition, Deletion: s.Deletion}\n")
	b.WriteString("\t}\n")
	b.WriteString("\tdata, err := json.Marshal(out)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*jsonOut = C.CString(string(data))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideCommitPatch(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(cHandle C.longlong, toHandle C.longlong, patchOut **C.char) *C.char {\n", cName)
	loadObjectReceiver(b, "Commit", "cHandle")
	b.WriteString("\tvar to *object.Commit\n")
	b.WriteString("\tif int64(toHandle) != 0 {\n")
	b.WriteString("\t\tvar ok bool\n")
	b.WriteString("\t\tto, ok = loadHandle[*object.Commit](int64(toHandle))\n")
	b.WriteString("\t\tif !ok {\n\t\t\treturn C.CString(\"invalid to-commit handle\")\n\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("\tpatch, err := recv.Patch(to)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*patchOut = C.CString(patch.String())\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideCommitMergeBase(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(cHandle C.longlong, otherHandle C.longlong, jsonOut **C.char) *C.char {\n", cName)
	loadObjectReceiver(b, "Commit", "cHandle")
	b.WriteString("\tother, ok := loadHandle[*object.Commit](int64(otherHandle))\n")
	b.WriteString("\tif !ok {\n\t\treturn C.CString(\"invalid other commit handle\")\n\t}\n")
	b.WriteString("\tbases, err := recv.MergeBase(other)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\thashes := make([]string, len(bases))\n")
	b.WriteString("\tfor i, c := range bases {\n")
	b.WriteString("\t\thashes[i] = c.Hash.String()\n")
	b.WriteString("\t}\n")
	b.WriteString("\tdata, err := json.Marshal(hashes)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*jsonOut = C.CString(string(data))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideCommitVerify(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(cHandle C.longlong, armoredKeyRing *C.char) *C.char {\n", cName)
	loadObjectReceiver(b, "Commit", "cHandle")
	b.WriteString("\t_, err := recv.Verify(C.GoString(armoredKeyRing))\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideTreeDiff(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(tHandle C.longlong, toHandle C.longlong, jsonOut **C.char) *C.char {\n", cName)
	loadObjectReceiver(b, "Tree", "tHandle")
	b.WriteString("\tvar to *object.Tree\n")
	b.WriteString("\tif int64(toHandle) != 0 {\n")
	b.WriteString("\t\tvar ok bool\n")
	b.WriteString("\t\tto, ok = loadHandle[*object.Tree](int64(toHandle))\n")
	b.WriteString("\t\tif !ok {\n\t\t\treturn C.CString(\"invalid to-tree handle\")\n\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("\tchanges, err := recv.Diff(to)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\ttype changeJSON struct {\n")
	b.WriteString("\t\tAction   string `json:\"action\"`\n")
	b.WriteString("\t\tFromPath string `json:\"fromPath,omitempty\"`\n")
	b.WriteString("\t\tToPath   string `json:\"toPath,omitempty\"`\n")
	b.WriteString("\t\tFromHash string `json:\"fromHash,omitempty\"`\n")
	b.WriteString("\t\tToHash   string `json:\"toHash,omitempty\"`\n")
	b.WriteString("\t}\n")
	b.WriteString("\tout := make([]changeJSON, len(changes))\n")
	b.WriteString("\tfor i, ch := range changes {\n")
	b.WriteString("\t\tc := changeJSON{}\n")
	b.WriteString("\t\tswitch {\n")
	b.WriteString("\t\tcase ch.From.Name == \"\" && ch.To.Name != \"\":\n")
	b.WriteString("\t\t\tc.Action = \"Insert\"\n")
	b.WriteString("\t\tcase ch.From.Name != \"\" && ch.To.Name == \"\":\n")
	b.WriteString("\t\t\tc.Action = \"Delete\"\n")
	b.WriteString("\t\tdefault:\n")
	b.WriteString("\t\t\tc.Action = \"Modify\"\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tif ch.From.Name != \"\" {\n")
	b.WriteString("\t\t\tc.FromPath = ch.From.Name\n")
	b.WriteString("\t\t\tc.FromHash = ch.From.TreeEntry.Hash.String()\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tif ch.To.Name != \"\" {\n")
	b.WriteString("\t\t\tc.ToPath = ch.To.Name\n")
	b.WriteString("\t\t\tc.ToHash = ch.To.TreeEntry.Hash.String()\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tout[i] = c\n")
	b.WriteString("\t}\n")
	b.WriteString("\tdata, err := json.Marshal(out)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*jsonOut = C.CString(string(data))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideTreePatch(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(tHandle C.longlong, toHandle C.longlong, patchOut **C.char) *C.char {\n", cName)
	loadObjectReceiver(b, "Tree", "tHandle")
	b.WriteString("\tvar to *object.Tree\n")
	b.WriteString("\tif int64(toHandle) != 0 {\n")
	b.WriteString("\t\tvar ok bool\n")
	b.WriteString("\t\tto, ok = loadHandle[*object.Tree](int64(toHandle))\n")
	b.WriteString("\t\tif !ok {\n\t\t\treturn C.CString(\"invalid to-tree handle\")\n\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("\tpatch, err := recv.Patch(to)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*patchOut = C.CString(patch.String())\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideTreeFindEntry(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(tHandle C.longlong, path *C.char, jsonOut **C.char) *C.char {\n", cName)
	loadObjectReceiver(b, "Tree", "tHandle")
	b.WriteString("\tentry, err := recv.FindEntry(C.GoString(path))\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\ttype entryJSON struct {\n")
	b.WriteString("\t\tName string `json:\"name\"`\n")
	b.WriteString("\t\tHash string `json:\"hash\"`\n")
	b.WriteString("\t\tMode uint32 `json:\"mode\"`\n")
	b.WriteString("\t}\n")
	b.WriteString("\tout := entryJSON{\n")
	b.WriteString("\t\tName: entry.Name,\n")
	b.WriteString("\t\tHash: entry.Hash.String(),\n")
	b.WriteString("\t\tMode: uint32(entry.Mode),\n")
	b.WriteString("\t}\n")
	b.WriteString("\tdata, err := json.Marshal(out)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*jsonOut = C.CString(string(data))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateOverrideTagVerify(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(tHandle C.longlong, armoredKeyRing *C.char) *C.char {\n", cName)
	loadObjectReceiver(b, "Tag", "tHandle")
	b.WriteString("\t_, err := recv.Verify(C.GoString(armoredKeyRing))\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\treturn nil\n}\n\n")
}

// --- New extra methods ---

func generateExtraRemoteConfig(b *strings.Builder) {
	b.WriteString(`//export GitRemoteConfig
func GitRemoteConfig(remoteHandle C.longlong, jsonOut **C.char) *C.char {
	remote, ok := loadHandle[*git.Remote](int64(remoteHandle))
	if !ok {
		return C.CString("invalid remote handle")
	}
	cfg := remote.Config()
	type remoteConfigJSON struct {
		Name  string   ` + "`json:\"name\"`" + `
		URLs  []string ` + "`json:\"urls\"`" + `
		Fetch []string ` + "`json:\"fetch\"`" + `
	}
	fetchSpecs := make([]string, len(cfg.Fetch))
	for i, f := range cfg.Fetch {
		fetchSpecs[i] = string(f)
	}
	out := remoteConfigJSON{Name: cfg.Name, URLs: cfg.URLs, Fetch: fetchSpecs}
	data, err := json.Marshal(out)
	if err != nil {
		return toCError(err)
	}
	*jsonOut = C.CString(string(data))
	return nil
}

`)
}

func generateExtraSubmoduleConfig(b *strings.Builder) {
	b.WriteString(`//export GitSubmoduleConfig
func GitSubmoduleConfig(subHandle C.longlong, jsonOut **C.char) *C.char {
	sub, ok := loadHandle[*git.Submodule](int64(subHandle))
	if !ok {
		return C.CString("invalid submodule handle")
	}
	cfg := sub.Config()
	type subConfigJSON struct {
		Name   string ` + "`json:\"name\"`" + `
		Path   string ` + "`json:\"path\"`" + `
		URL    string ` + "`json:\"url\"`" + `
		Branch string ` + "`json:\"branch\"`" + `
	}
	out := subConfigJSON{Name: cfg.Name, Path: cfg.Path, URL: cfg.URL, Branch: string(cfg.Branch)}
	data, err := json.Marshal(out)
	if err != nil {
		return toCError(err)
	}
	*jsonOut = C.CString(string(data))
	return nil
}

`)
}

func generateExtraSubmoduleStatus(b *strings.Builder) {
	b.WriteString(`//export GitSubmoduleStatus
func GitSubmoduleStatus(subHandle C.longlong, jsonOut **C.char) *C.char {
	sub, ok := loadHandle[*git.Submodule](int64(subHandle))
	if !ok {
		return C.CString("invalid submodule handle")
	}
	status, err := sub.Status()
	if err != nil {
		return toCError(err)
	}
	type statusJSON struct {
		Path     string ` + "`json:\"path\"`" + `
		Current  string ` + "`json:\"current\"`" + `
		Expected string ` + "`json:\"expected\"`" + `
		Branch   string ` + "`json:\"branch\"`" + `
	}
	out := statusJSON{
		Path:     status.Path,
		Current:  status.Current.String(),
		Expected: status.Expected.String(),
		Branch:   string(status.Branch),
	}
	data, err := json.Marshal(out)
	if err != nil {
		return toCError(err)
	}
	*jsonOut = C.CString(string(data))
	return nil
}

`)
}

func generateExtraBlame(b *strings.Builder) {
	b.WriteString(`//export GitBlame
func GitBlame(commitHandle C.longlong, path *C.char, jsonOut **C.char) *C.char {
	commit, ok := loadHandle[*object.Commit](int64(commitHandle))
	if !ok {
		return C.CString("invalid commit handle")
	}
	result, err := git.Blame(commit, C.GoString(path))
	if err != nil {
		return toCError(err)
	}
	type lineJSON struct {
		Author      string ` + "`json:\"author\"`" + `
		AuthorEmail string ` + "`json:\"authorEmail\"`" + `
		Hash        string ` + "`json:\"hash\"`" + `
		Date        int64  ` + "`json:\"date\"`" + `
		Text        string ` + "`json:\"text\"`" + `
	}
	type blameJSON struct {
		Path  string     ` + "`json:\"path\"`" + `
		Rev   string     ` + "`json:\"rev\"`" + `
		Lines []lineJSON ` + "`json:\"lines\"`" + `
	}
	out := blameJSON{Path: result.Path, Rev: result.Rev.String()}
	for _, l := range result.Lines {
		out.Lines = append(out.Lines, lineJSON{
			Author:      l.AuthorName,
			AuthorEmail: l.Author,
			Hash:        l.Hash.String(),
			Date:        l.Date.Unix(),
			Text:        l.Text,
		})
	}
	data, err := json.Marshal(out)
	if err != nil {
		return toCError(err)
	}
	*jsonOut = C.CString(string(data))
	return nil
}

`)
}

func generateExtraCommitFieldAccessors(b *strings.Builder) {
	for _, field := range []struct{ name, accessor string }{
		{"Author", "Author"},
		{"Committer", "Committer"},
	} {
		for _, prop := range []struct{ name, expr, cType string }{
			{"Name", field.accessor + ".Name", "**C.char"},
			{"Email", field.accessor + ".Email", "**C.char"},
		} {
			exportName := fmt.Sprintf("GitCommit%s%s", field.name, prop.name)
			fmt.Fprintf(b, "//export %s\n", exportName)
			fmt.Fprintf(b, "func %s(cHandle C.longlong, out %s) *C.char {\n", exportName, prop.cType)
			fmt.Fprintf(b, "\trecv, ok := loadHandle[*object.Commit](int64(cHandle))\n")
			fmt.Fprintf(b, "\tif !ok {\n\t\treturn C.CString(\"invalid commit handle\")\n\t}\n")
			fmt.Fprintf(b, "\t*out = C.CString(recv.%s)\n", prop.expr)
			fmt.Fprintf(b, "\treturn nil\n}\n\n")
		}
		exportName := fmt.Sprintf("GitCommit%sWhen", field.name)
		fmt.Fprintf(b, "//export %s\n", exportName)
		fmt.Fprintf(b, "func %s(cHandle C.longlong, tsOut *C.longlong) *C.char {\n", exportName)
		fmt.Fprintf(b, "\trecv, ok := loadHandle[*object.Commit](int64(cHandle))\n")
		fmt.Fprintf(b, "\tif !ok {\n\t\treturn C.CString(\"invalid commit handle\")\n\t}\n")
		fmt.Fprintf(b, "\t*tsOut = C.longlong(recv.%s.When.Unix())\n", field.accessor)
		fmt.Fprintf(b, "\treturn nil\n}\n\n")
	}
}

