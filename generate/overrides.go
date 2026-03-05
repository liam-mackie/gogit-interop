package main

import (
	"fmt"
	"strings"
)

var overrideMethods = map[string]map[string]bool{
	"Repository": {
		"CreateRemote": true, "CreateBranch": true,
		"CommitObject": true, "Merge": true,
	},
	"Worktree": {
		"Status": true, "Submodules": true,
	},
	"Remote": {
		"List": true,
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
	case "Repository.CommitObject":
		generateOverrideCommitObject(b, m.CName)
	case "Repository.Merge":
		generateOverrideMerge(b, m.CName)
	case "Worktree.Status":
		generateOverrideWorktreeStatus(b, m.CName)
	case "Worktree.Submodules":
		generateOverrideWorktreeSubmodules(b, m.CName)
	case "Remote.List":
		generateOverrideRemoteList(b, m.CName)
	}
}

func generateExtraMethodsGo(b *strings.Builder, ht *HandleType) {
	switch ht.GoName {
	case "Repository":
		generateExtraRepoRemotes(b)
	case "Remote":
		generateExtraRemoteConfigName(b)
	case "Submodule":
		generateExtraSubmoduleConfigName(b)
	}
}

func writeOverrideNativeMethod(b *strings.Builder, ht *HandleType, m Method) {
	cName := m.CName
	switch ht.GoName + "." + m.GoName {
	case "Repository.CreateRemote":
		writeDllImport(b, cName, "long repoHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string name,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string url,\n        out long handleOut")
	case "Repository.CreateBranch":
		writeDllImport(b, cName, "long repoHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string name,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string hash")
	case "Repository.CommitObject":
		writeDllImport(b, cName, "long repoHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string hash,\n        out IntPtr commitHashOut,\n        out IntPtr msgOut,\n        out IntPtr authorNameOut,\n        out IntPtr authorEmailOut,\n        out long tsOut")
	case "Repository.Merge":
		writeDllImport(b, cName, "long repoHandle,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string refName,\n        [MarshalAs(UnmanagedType.LPUTF8Str)] string hash,\n        long optsHandle")
	case "Worktree.Status":
		writeDllImport(b, cName, "long wtHandle, out IntPtr jsonOut")
	case "Worktree.Submodules":
		writeDllImport(b, cName, "long wtHandle, out IntPtr jsonOut")
	case "Remote.List":
		writeDllImport(b, cName, "long remoteHandle, long optsHandle, out IntPtr jsonOut")
	default:
		writeGenericNativeMethod(b, ht, m)
	}
}

func writeExtraNativeMethods(b *strings.Builder, ht *HandleType) {
	switch ht.GoName {
	case "Repository":
		writeDllImport(b, "GitRepositoryRemotes", "long repoHandle, out IntPtr jsonOut")
	case "Remote":
		writeDllImport(b, "GitRemoteConfigName", "long remoteHandle, out IntPtr nameOut")
	case "Submodule":
		writeDllImport(b, "GitSubmoduleConfigName", "long subHandle, out IntPtr nameOut")
	}
}

// --- Go override implementations ---

func loadReceiver(b *strings.Builder, typeName, handleParam string) {
	goType := fmt.Sprintf("*git.%s", typeName)
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

func generateOverrideCommitObject(b *strings.Builder, cName string) {
	fmt.Fprintf(b, "//export %s\n", cName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, hash *C.char, commitHashOut **C.char, msgOut **C.char, authorNameOut **C.char, authorEmailOut **C.char, tsOut *C.longlong) *C.char {\n", cName)
	loadReceiver(b, "Repository", "repoHandle")
	b.WriteString("\th := plumbing.NewHash(C.GoString(hash))\n")
	b.WriteString("\tcommit, err := recv.CommitObject(h)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*commitHashOut = C.CString(commit.Hash.String())\n")
	b.WriteString("\t*msgOut = C.CString(commit.Message)\n")
	b.WriteString("\t*authorNameOut = C.CString(commit.Author.Name)\n")
	b.WriteString("\t*authorEmailOut = C.CString(commit.Author.Email)\n")
	b.WriteString("\t*tsOut = C.longlong(commit.Author.When.Unix())\n")
	b.WriteString("\treturn nil\n}\n\n")
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
