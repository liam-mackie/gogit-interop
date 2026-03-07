package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateCSharpWrappers(pkg *Package, dir string) error {
	for _, ht := range pkg.Types {
		if err := generateCSharpWrapper(pkg, &ht, dir); err != nil {
			return fmt.Errorf("generating %s.cs: %w", ht.GoName, err)
		}
	}
	return nil
}

func generateCSharpWrapper(pkg *Package, ht *HandleType, dir string) error {
	var b strings.Builder

	needsJson := wrapperNeedsJson(ht)
	b.WriteString(csGenHeader)
	b.WriteString("#nullable enable\n")
	if needsJson {
		b.WriteString("using System.Text.Json;\n")
	}
	b.WriteString("\nnamespace GoGitDotNet;\n\n")
	if doc := csTypeDoc(ht.GoName); doc != "" {
		fmt.Fprintf(&b, "/// <summary>%s</summary>\n", doc)
	}
	fmt.Fprintf(&b, "public sealed class %s : IDisposable\n{\n", ht.GoName)
	b.WriteString("    private long _handle;\n")
	b.WriteString("    private bool _disposed;\n\n")
	fmt.Fprintf(&b, "    internal %s(long handle) => _handle = handle;\n", ht.GoName)
	b.WriteString("    internal long Handle => _handle;\n")

	if ht.GoName == "Repository" {
		generateRepoFactoryMethods(&b, pkg)
	}

	for _, m := range ht.Methods {
		b.WriteString("\n")
		if isOverrideMethod(ht.GoName, m.GoName) {
			generateOverrideWrapperMethod(&b, ht, m)
		} else {
			generateGenericWrapperMethod(&b, ht, m)
		}
	}

	generateExtraWrapperMethods(&b, ht)
	generateFieldProperties(&b, ht)

	b.WriteString("\n    public void Dispose()\n    {\n")
	b.WriteString("        if (_disposed) return;\n")
	b.WriteString("        _disposed = true;\n")
	fmt.Fprintf(&b, "        NativeMethods.%sFree(_handle);\n", ht.CPrefix)
	b.WriteString("        _handle = 0;\n")
	b.WriteString("    }\n}\n")

	return os.WriteFile(filepath.Join(dir, ht.GoName+".cs"), []byte(b.String()), 0644)
}

func wrapperNeedsJson(ht *HandleType) bool {
	switch ht.GoName {
	case "Repository", "Worktree", "Remote", "Commit", "Tree", "Submodule":
		return true
	}
	for _, m := range ht.Methods {
		for _, r := range m.Returns {
			if r.Mapping.Kind == MappingStringSlice {
				return true
			}
		}
	}
	return false
}

func generateRepoFactoryMethods(b *strings.Builder, pkg *Package) {
	b.WriteString(`
    /// <summary>Initialises a new git repository at <paramref name="path"/>.</summary>
    public static Repository Init(string path, bool isBare = false)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitPlainInit(path, isBare ? 1 : 0, out var handle));
        return new Repository(handle);
    }

    /// <summary>Opens an existing git repository at <paramref name="path"/>.</summary>
    public static Repository Open(string path)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitPlainOpen(path, out var handle));
        return new Repository(handle);
    }

    /// <summary>Opens an existing git repository at <paramref name="path"/> with additional options.</summary>
    public static Repository OpenWithOptions(string path, PlainOpenOptions options)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitPlainOpenWithOptions(path, options.Handle, out var handle));
        return new Repository(handle);
    }

    /// <summary>Clones the repository described by <paramref name="options"/> into <paramref name="path"/> on disk.</summary>
    public static Repository Clone(string path, CloneOptions options)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitPlainClone(path, options.Handle, out var handle));
        return new Repository(handle);
    }

    /// <summary>Clones the repository described by <paramref name="options"/> entirely into memory. No files are written to disk.</summary>
    public static Repository CloneInMemory(CloneOptions options)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitCloneInMemory(options.Handle, out var handle));
        return new Repository(handle);
    }
`)
}

func generateGenericWrapperMethod(b *strings.Builder, ht *HandleType, m Method) {
	returnType := csWrapperReturnType(m)
	publicName := csPublicMethodName(ht, m)
	params := csWrapperParams(m)

	if doc := csMethodDoc(ht.GoName, m.GoName); doc != "" {
		fmt.Fprintf(b, "    /// <summary>%s</summary>\n", doc)
	}
	fmt.Fprintf(b, "    public %s %s(%s)\n    {\n", returnType, publicName, params)
	b.WriteString("        ObjectDisposedException.ThrowIf(_disposed, this);\n")

	generateWrapperBody(b, ht, m)

	b.WriteString("    }\n")
}

func generateWrapperBody(b *strings.Builder, ht *HandleType, m Method) {
	hasNonErrorReturn := false
	for _, r := range m.Returns {
		if !r.IsError {
			hasNonErrorReturn = true
			break
		}
	}

	nativeArgs := csWrapperNativeArgs(m)

	if !hasNonErrorReturn {
		fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s));\n", m.CName, nativeArgs)
		return
	}

	for _, r := range m.Returns {
		if r.IsError {
			continue
		}
		switch r.Mapping.Kind {
		case MappingReference:
			fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s, out var outRefName, out var outHash));\n", m.CName, nativeArgs)
			b.WriteString("        return (NativeMethods.ConsumeGoString(outRefName)!, NativeMethods.ConsumeGoString(outHash)!);\n")
		case MappingHandle:
			fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s, out var resultHandle));\n", m.CName, nativeArgs)
			fmt.Fprintf(b, "        return new %s(resultHandle);\n", r.Mapping.HandleType)
		case MappingIterator:
			fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s, out var iter));\n", m.CName, nativeArgs)
			fmt.Fprintf(b, "        return new %s(iter);\n", csIteratorClassName(r.Mapping.HandleType))
		case MappingHash:
			fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s, out var hash));\n", m.CName, nativeArgs)
			b.WriteString("        return NativeMethods.ConsumeGoString(hash)!;\n")
		case MappingString, MappingReferenceName, MappingRevision:
			fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s, out var s));\n", m.CName, nativeArgs)
			b.WriteString("        return NativeMethods.ConsumeGoString(s)!;\n")
		case MappingBool:
			fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s, out var val));\n", m.CName, nativeArgs)
			b.WriteString("        return val != 0;\n")
		case MappingPrimitive, MappingEnum:
			fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s, out var val));\n", m.CName, nativeArgs)
			b.WriteString("        return val;\n")
		case MappingTime:
			fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s, out var val));\n", m.CName, nativeArgs)
			b.WriteString("        return DateTimeOffset.FromUnixTimeSeconds(val);\n")
		case MappingStringSlice:
			fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s, out var jsonPtr));\n", m.CName, nativeArgs)
			b.WriteString("        var json = NativeMethods.ConsumeGoString(jsonPtr)!;\n")
			b.WriteString("        return JsonSerializer.Deserialize<string[]>(json) ?? [];\n")
		default:
			fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s));\n", m.CName, nativeArgs)
		}
		break
	}
}

func csWrapperReturnType(m Method) string {
	for _, r := range m.Returns {
		if r.IsError {
			continue
		}
		switch r.Mapping.Kind {
		case MappingReference:
			return "(string RefName, string Hash)"
		case MappingHandle:
			return r.Mapping.HandleType
		case MappingIterator:
			return csIteratorClassName(r.Mapping.HandleType)
		case MappingHash, MappingString, MappingReferenceName, MappingRevision:
			return "string"
		case MappingBool:
			return "bool"
		case MappingPrimitive, MappingEnum:
			return r.Mapping.CSharpType
		case MappingTime:
			return "DateTimeOffset"
		case MappingStringSlice:
			return "string[]"
		}
	}
	return "void"
}

func csPublicMethodName(ht *HandleType, m Method) string {
	switch ht.GoName + "." + m.GoName {
	case "Repository.Worktree":
		return "GetWorktree"
	case "Repository.Remote":
		return "GetRemote"
	case "Repository.CommitObject":
		return "GetCommitObject"
	case "Repository.TreeObject":
		return "GetTreeObject"
	case "Repository.BlobObject":
		return "GetBlobObject"
	case "Repository.TagObject":
		return "GetTagObject"
	case "Worktree.Submodule":
		return "GetSubmodule"
	case "Submodule.Repository":
		return "GetRepository"
	case "Tree.Tree":
		return "GetSubtree"
	case "Commit.Tree":
		return "GetTree"
	case "Tag.Commit":
		return "GetCommit"
	case "Tag.Tree":
		return "GetTree"
	case "Tag.Blob":
		return "GetBlob"
	}
	return m.GoName
}

func csWrapperParams(m Method) string {
	var parts []string
	for _, p := range m.Params {
		csType := csWrapperParamType(p)
		name := csParamName(p)
		if p.IsOptional {
			parts = append(parts, fmt.Sprintf("%s? %s = null", csType, name))
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", csType, name))
		}
	}
	return strings.Join(parts, ", ")
}

func csWrapperParamType(p Param) string {
	switch p.Mapping.Kind {
	case MappingString, MappingReferenceName, MappingHash, MappingRevision:
		return "string"
	case MappingBool:
		return "bool"
	case MappingPrimitive:
		return p.Mapping.CSharpType
	case MappingEnum:
		return p.Mapping.CSharpType
	case MappingOptions:
		return p.Mapping.HandleType
	case MappingHandle:
		return p.Mapping.HandleType
	case MappingAuth:
		return "Auth"
	case MappingSigner:
		return "Signer"
	case MappingTime:
		return "DateTimeOffset"
	case MappingDuration:
		return "TimeSpan"
	case MappingStringSlice:
		return "string[]"
	default:
		return "long"
	}
}

func csIteratorClassName(handleType string) string {
	switch handleType {
	case "CommitIter":
		return "CommitIterator"
	case "ReferenceIter":
		return "ReferenceIterator"
	case "FileIter":
		return "FileIterator"
	case "TreeIter":
		return "TreeIterator"
	case "BlobIter":
		return "BlobIterator"
	case "TagIter":
		return "TagIterator"
	default:
		return handleType + "Iterator"
	}
}

func generateFieldProperties(b *strings.Builder, ht *HandleType) {
	for _, f := range ht.Fields {
		csType := csFieldPropertyType(f)
		if csType == "" {
			continue
		}

		if doc := csFieldDoc(ht.GoName, f.GoName); doc != "" {
			fmt.Fprintf(b, "\n    /// <summary>%s</summary>", doc)
		}
		fmt.Fprintf(b, "\n    public %s %s\n    {\n", csType, f.GoName)
		b.WriteString("        get\n        {\n")
		b.WriteString("            ObjectDisposedException.ThrowIf(_disposed, this);\n")

		switch f.Mapping.Kind {
		case MappingString, MappingHash, MappingReferenceName:
			fmt.Fprintf(b, "            NativeMethods.ThrowIfError(NativeMethods.%s(_handle, out var ptr));\n", f.CGetterName)
			b.WriteString("            return NativeMethods.ConsumeGoString(ptr)!;\n")
		case MappingPrimitive:
			fmt.Fprintf(b, "            NativeMethods.ThrowIfError(NativeMethods.%s(_handle, out var val));\n", f.CGetterName)
			b.WriteString("            return val;\n")
		case MappingBool:
			fmt.Fprintf(b, "            NativeMethods.ThrowIfError(NativeMethods.%s(_handle, out var val));\n", f.CGetterName)
			b.WriteString("            return val != 0;\n")
		}

		b.WriteString("        }\n    }\n")
	}
}

func csFieldPropertyType(f HandleField) string {
	switch f.Mapping.Kind {
	case MappingString, MappingHash, MappingReferenceName:
		return "string"
	case MappingPrimitive:
		return f.Mapping.CSharpType
	case MappingBool:
		return "bool"
	default:
		return ""
	}
}

func csTypeDoc(typeName string) string {
	switch typeName {
	case "Repository":
		return "A git repository. Provides access to commits, branches, tags, remotes, and worktree operations. Wraps <c>*git.Repository</c> from go-git."
	case "Remote":
		return "A git remote. Use <see cref=\"Create\"/> to connect to a remote URL without a local clone, or obtain one from a <see cref=\"Repository\"/>. Wraps <c>*git.Remote</c> from go-git."
	case "Worktree":
		return "The working tree of a repository. Provides staging, committing, checkout, reset, restore, and status operations. Wraps <c>*git.Worktree</c> from go-git."
	case "Commit":
		return "A git commit object. Provides access to the hash, message, author, committer, tree, parents, and diff operations. Wraps <c>*object.Commit</c> from go-git."
	case "Tree":
		return "A git tree object representing a directory snapshot. Provides file enumeration, entry lookup, and diff operations. Wraps <c>*object.Tree</c> from go-git."
	case "Blob":
		return "A git blob object representing raw file contents. Wraps <c>*object.Blob</c> from go-git."
	case "Tag":
		return "A git tag object. Provides access to tag name, message, tagger, and the tagged object. Wraps <c>*object.Tag</c> from go-git."
	case "File":
		return "A file inside a git tree. Provides access to the file name, hash, and contents. Wraps <c>*object.File</c> from go-git."
	case "Submodule":
		return "A git submodule. Provides init, update, config, and status operations. Wraps <c>*git.Submodule</c> from go-git."
	}
	return ""
}

func csMethodDoc(typeName, methodName string) string {
	key := typeName + "." + methodName
	docs := map[string]string{
		// Repository
		"Repository.Fetch":                   "Fetches from all configured remotes.",
		"Repository.FetchContext":             "Fetches from all configured remotes using a background context.",
		"Repository.Push":                     "Pushes local changes to the remote.",
		"Repository.PushContext":              "Pushes local changes to the remote using a background context.",
		"Repository.Head":                     "Returns the HEAD reference as a (refName, hash) pair.",
		"Repository.Log":                      "Returns an iterator over commits reachable from the options starting point.",
		"Repository.References":               "Returns an iterator over all references in the repository.",
		"Repository.Reference":                "Looks up a single reference by name.",
		"Repository.ResolveRevision":          "Resolves a revision string (e.g. <c>HEAD~2</c>) to a commit hash.",
		"Repository.Branches":                 "Returns an iterator over all branch references.",
		"Repository.Tags":                     "Returns an iterator over all tag references.",
		"Repository.Notes":                    "Returns an iterator over all note references.",
		"Repository.CommitObjects":            "Returns an iterator over all commit objects in the object store.",
		"Repository.TreeObjects":              "Returns an iterator over all tree objects in the object store.",
		"Repository.BlobObjects":              "Returns an iterator over all blob objects in the object store.",
		"Repository.TagObjects":               "Returns an iterator over all annotated tag objects in the object store.",
		"Repository.DeleteBranch":             "Deletes a local branch by name.",
		"Repository.DeleteTag":                "Deletes a tag by name.",
		"Repository.DeleteRemote":             "Removes a remote configuration by name.",
		"Repository.DeleteObject":             "Removes an object from the object store by hash.",
		"Repository.Merge":                    "Merges the given commit into the current branch.",
		// Remote
		"Remote.Fetch":                        "Fetches from this remote.",
		"Remote.FetchContext":                 "Fetches from this remote using a background context.",
		"Remote.Push":                         "Pushes to this remote.",
		"Remote.PushContext":                  "Pushes to this remote using a background context.",
		"Remote.String":                       "Returns a human-readable description of the remote.",
		"Remote.List":                         "Lists references advertised by the remote server. Does not require a local clone.",
		// Worktree
		"Worktree.Add":                        "Stages a file at the given path.",
		"Worktree.Commit":                     "Creates a new commit with the staged changes.",
		"Worktree.Checkout":                   "Checks out a branch, tag, or commit.",
		"Worktree.Reset":                      "Resets the working tree and/or index to the given state.",
		"Worktree.Restore":                    "Restores working tree files.",
		"Worktree.Clean":                      "Removes untracked files from the working tree.",
		"Worktree.Pull":                       "Fetches and merges from the tracked remote branch.",
		"Worktree.PullContext":                "Fetches and merges from the tracked remote branch using a background context.",
		"Worktree.Grep":                       "Searches working tree files matching the given options.",
		// Commit
		"Commit.Tree":                         "Returns the root tree of this commit.",
		"Commit.Parents":                      "Returns an iterator over the parent commits.",
		"Commit.Files":                        "Returns an iterator over all files in the commit's tree.",
		"Commit.Stats":                        "Returns per-file line addition/deletion statistics for this commit.",
		"Commit.Patch":                        "Returns the unified diff between this commit and the given commit (or the empty tree if nil).",
		"Commit.MergeBase":                    "Returns the common ancestor commits of this commit and the given commit.",
		"Commit.Verify":                       "Verifies the PGP signature on this commit against the provided armored key ring.",
		// Tree
		"Tree.Files":                          "Returns an iterator over all files reachable from this tree.",
		"Tree.Entries":                        "Returns the direct entries of this tree.",
		"Tree.TreeEntries":                    "Returns the direct subtree entries of this tree.",
		"Tree.Diff":                           "Returns the list of changes between this tree and the given tree (or the empty tree if nil).",
		"Tree.Patch":                          "Returns the unified diff between this tree and the given tree (or the empty tree if nil).",
		"Tree.FindEntry":                      "Finds a tree entry by path, searching recursively.",
		// Tag
		"Tag.Commit":                          "Returns the commit pointed to by this tag.",
		"Tag.Tree":                            "Returns the tree pointed to by this tag.",
		"Tag.Blob":                            "Returns the blob pointed to by this tag.",
		"Tag.Object":                          "Returns the object pointed to by this tag.",
		"Tag.Verify":                          "Verifies the PGP signature on this tag against the provided armored key ring.",
		// Blob
		"Blob.Reader":                         "Returns the raw byte contents of this blob.",
		// Submodule
		"Submodule.Init":                      "Initialises the submodule by registering its URL in the local git config.",
		"Submodule.Update":                    "Fetches and checks out the submodule at the commit recorded in the parent repo.",
		"Submodule.Repository":                "Returns the Repository for this submodule.",
	}
	if doc, ok := docs[key]; ok {
		return doc
	}
	return ""
}

func csFieldDoc(typeName, fieldName string) string {
	key := typeName + "." + fieldName
	docs := map[string]string{
		"Commit.Hash":        "The SHA-1 hash of this commit.",
		"Commit.Message":     "The full commit message.",
		"Tree.Hash":          "The SHA-1 hash of this tree.",
		"Blob.Hash":          "The SHA-1 hash of this blob.",
		"Blob.Size":          "The size of the blob contents in bytes.",
		"Tag.Hash":           "The SHA-1 hash of this tag object.",
		"Tag.Name":           "The tag name.",
		"Tag.Message":        "The tag message.",
		"Tag.TargetType":     "The object type of the tagged target.",
		"Tag.Target":         "The SHA-1 hash of the tagged object.",
		"File.Name":          "The file path relative to the repository root.",
		"File.Hash":          "The SHA-1 hash of this file's blob.",
	}
	if doc, ok := docs[key]; ok {
		return doc
	}
	return ""
}

var csReservedWords = map[string]string{
	"in": "input", "out": "output", "ref": "reference",
	"string": "str", "object": "obj", "event": "evt",
	"base": "baseVal", "params": "parameters",
}

func csParamName(p Param) string {
	name := p.GoName
	if name == "" {
		name = p.CName
	}
	name = strings.TrimSuffix(name, "Handle")
	name = lowerFirst(name)
	if replacement, ok := csReservedWords[name]; ok {
		return replacement
	}
	return name
}

func csWrapperNativeArgs(m Method) string {
	var parts []string
	for _, p := range m.Params {
		parts = append(parts, csWrapperMarshalArg(p))
	}
	if len(parts) == 0 {
		return ""
	}
	return ", " + strings.Join(parts, ", ")
}

func csWrapperMarshalArg(p Param) string {
	name := csParamName(p)
	switch p.Mapping.Kind {
	case MappingString, MappingReferenceName, MappingHash, MappingRevision, MappingStringSlice:
		return name
	case MappingBool:
		return name + " ? 1 : 0"
	case MappingOptions:
		if p.IsOptional {
			return name + "?.Handle ?? 0"
		}
		return name + ".Handle"
	case MappingHandle:
		return name + ".Handle"
	case MappingAuth:
		return name + ".Handle"
	case MappingSigner:
		return name + ".Handle"
	case MappingTime:
		return name + ".ToUnixTimeSeconds()"
	case MappingDuration:
		return name + ".Ticks * 100"
	default:
		return name
	}
}

func generateOverrideWrapperMethod(b *strings.Builder, ht *HandleType, m Method) {
	switch ht.GoName + "." + m.GoName {
	case "Repository.CreateRemote":
		b.WriteString(`    /// <summary>Adds a new remote with the given <paramref name="name"/> and <paramref name="url"/> to the repository configuration.</summary>
    public Remote CreateRemote(string name, string url)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, name, url, out var handle));
        return new Remote(handle);
    }
`)
	case "Repository.CreateBranch":
		b.WriteString(`    /// <summary>Creates a new local branch pointing at <paramref name="hash"/>.</summary>
    public void CreateBranch(string name, string hash)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, name, hash));
    }
`)
	case "Repository.CommitObject":
		b.WriteString(`    /// <summary>Looks up a commit object by its SHA-1 <paramref name="hash"/>.</summary>
    public Commit GetCommitObject(string hash)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(
            _handle, hash,
            out var commitHash, out var msg, out var authorName, out var authorEmail, out var ts));
        return new Commit
        {
            Hash = NativeMethods.ConsumeGoString(commitHash)!,
            Message = NativeMethods.ConsumeGoString(msg)!,
            AuthorName = NativeMethods.ConsumeGoString(authorName)!,
            AuthorEmail = NativeMethods.ConsumeGoString(authorEmail)!,
            AuthorTimestamp = DateTimeOffset.FromUnixTimeSeconds(ts),
        };
    }
`)
	case "Repository.Merge":
		b.WriteString(`    /// <summary>Merges the commit identified by <paramref name="hash"/> into the current branch.</summary>
    public void Merge(string refName, string hash, MergeOptions? options = null)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        var optsHandle = options?.Handle ?? 0;
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, refName, hash, optsHandle));
    }
`)
	case "Worktree.Status":
		b.WriteString(`    /// <summary>Returns the status of each file in the working tree, keyed by path.</summary>
    public Dictionary<string, FileStatus> Status()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<Dictionary<string, FileStatus>>(json) ?? new();
    }
`)
	case "Worktree.Submodules":
		b.WriteString(`    /// <summary>Returns the names of all submodules registered in this worktree.</summary>
    public string[] Submodules()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<string[]>(json) ?? [];
    }
`)
	case "Remote.List":
		b.WriteString(`    /// <summary>Lists all references advertised by the remote server. Does not require a local clone.</summary>
    public ReferenceInfo[] List(ListOptions options)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, options.Handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<ReferenceInfo[]>(json) ?? [];
    }
`)
	case "Blob.Reader":
		b.WriteString(`    /// <summary>Returns the full contents of this blob as a UTF-8 string.</summary>
    public string Contents()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, out var dataPtr));
        return NativeMethods.ConsumeGoString(dataPtr)!;
    }
`)
	case "Repository.Branch":
		b.WriteString(`    /// <summary>Returns the configuration for the branch named <paramref name="name"/>.</summary>
    public BranchConfig GetBranch(string name)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, name, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<BranchConfig>(json)!;
    }
`)
	case "Repository.Config":
		b.WriteString(`    /// <summary>Returns a subset of the repository's git configuration (core, user identity, default branch).</summary>
    public GitConfig GetConfig()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<GitConfig>(json)!;
    }
`)
	case "Repository.CreateRemoteAnonymous":
		b.WriteString(`    /// <summary>Creates a temporary anonymous remote for <paramref name="url"/>. The remote is not saved to the repository configuration.</summary>
    public Remote CreateRemoteAnonymous(string url)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, url, out var handle));
        return new Remote(handle);
    }
`)
	case "Commit.Stats":
		b.WriteString(`    /// <summary>Returns per-file line addition/deletion statistics for this commit.</summary>
    public FileStat[] Stats()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<FileStat[]>(json) ?? [];
    }
`)
	case "Commit.Patch":
		b.WriteString(`    /// <summary>Returns the unified diff between this commit and <paramref name="to"/>. Pass <c>null</c> to diff against the empty tree.</summary>
    public string GetPatch(Commit? to = null)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, to?.Handle ?? 0, out var patchPtr));
        return NativeMethods.ConsumeGoString(patchPtr)!;
    }
`)
	case "Commit.MergeBase":
		b.WriteString(`    /// <summary>Returns the SHA-1 hashes of common ancestor commits shared by this commit and <paramref name="other"/>.</summary>
    public string[] MergeBase(Commit other)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, other.Handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<string[]>(json) ?? [];
    }
`)
	case "Commit.Verify":
		b.WriteString(`    /// <summary>Verifies the PGP signature of this commit against <paramref name="armoredKeyRing"/>. Throws if the signature is invalid or missing.</summary>
    public void Verify(string armoredKeyRing)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, armoredKeyRing));
    }
`)
	case "Tree.Diff":
		b.WriteString(`    /// <summary>Returns the list of file changes between this tree and <paramref name="to"/>. Pass <c>null</c> to diff against the empty tree.</summary>
    public DiffChange[] Diff(Tree? to = null)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, to?.Handle ?? 0, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<DiffChange[]>(json) ?? [];
    }
`)
	case "Tree.Patch":
		b.WriteString(`    /// <summary>Returns the unified diff between this tree and <paramref name="to"/>. Pass <c>null</c> to diff against the empty tree.</summary>
    public string GetPatch(Tree? to = null)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, to?.Handle ?? 0, out var patchPtr));
        return NativeMethods.ConsumeGoString(patchPtr)!;
    }
`)
	case "Tree.FindEntry":
		b.WriteString(`    /// <summary>Finds the tree entry at the given <paramref name="path"/>, searching recursively through subtrees.</summary>
    public TreeEntryInfo FindEntry(string path)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, path, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<TreeEntryInfo>(json)!;
    }
`)
	case "Tag.Verify":
		b.WriteString(`    /// <summary>Verifies the PGP signature of this tag against <paramref name="armoredKeyRing"/>. Throws if the signature is invalid or missing.</summary>
    public void Verify(string armoredKeyRing)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, armoredKeyRing));
    }
`)
	}
}

func generateExtraWrapperMethods(b *strings.Builder, ht *HandleType) {
	switch ht.GoName {
	case "Repository":
		b.WriteString(`
    /// <summary>Returns the names of all configured remotes.</summary>
    public string[] Remotes()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.GitRepositoryRemotes(_handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<string[]>(json) ?? [];
    }

    /// <summary>Returns line-by-line blame information for the file at <paramref name="path"/> as of the given <paramref name="commit"/>.</summary>
    public static BlameResult Blame(Commit commit, string path)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitBlame(commit.Handle, path, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<BlameResult>(json)!;
    }

    /// <summary>Writes <paramref name="content"/> as a blob object to the object store. Returns its SHA-1 hash.</summary>
    public string StoreBlob(byte[] content)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        var b64 = Convert.ToBase64String(content);
        NativeMethods.ThrowIfError(NativeMethods.GitRepositoryStoreBlob(_handle, b64, out var hashPtr));
        return NativeMethods.ConsumeGoString(hashPtr)!;
    }

    /// <summary>Writes UTF-8 encoded <paramref name="text"/> as a blob object to the object store. Returns its SHA-1 hash.</summary>
    public string StoreBlob(string text) => StoreBlob(System.Text.Encoding.UTF8.GetBytes(text));

    /// <summary>Returns the direct entries of the tree identified by <paramref name="treeHash"/>.</summary>
    public TreeEntryInfo[] GetTreeEntries(string treeHash)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.GitRepositoryGetTreeEntries(_handle, treeHash, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<TreeEntryInfo[]>(json) ?? [];
    }

    /// <summary>Writes a tree object built from <paramref name="entries"/> to the object store. Returns its SHA-1 hash.</summary>
    public string StoreTree(IEnumerable<TreeEntryInfo> entries)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        var json = JsonSerializer.Serialize(entries);
        NativeMethods.ThrowIfError(NativeMethods.GitRepositoryStoreTree(_handle, json, out var hashPtr));
        return NativeMethods.ConsumeGoString(hashPtr)!;
    }

    /// <summary>
    /// Writes a commit object directly to the object store without touching the worktree.
    /// Returns the SHA-1 hash of the new commit.
    /// </summary>
    public string StoreCommit(
        string treeHash,
        string[] parentHashes,
        string authorName, string authorEmail,
        string committerName, string committerEmail,
        string message,
        DateTimeOffset when)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        var parentsJson = JsonSerializer.Serialize(parentHashes);
        var ts = when.ToUnixTimeSeconds();
        NativeMethods.ThrowIfError(NativeMethods.GitRepositoryStoreCommit(
            _handle, treeHash, parentsJson,
            authorName, authorEmail,
            committerName, committerEmail,
            message, ts, out var hashPtr));
        return NativeMethods.ConsumeGoString(hashPtr)!;
    }

    /// <summary>Updates (or creates) the reference <paramref name="refName"/> to point at <paramref name="hash"/>.</summary>
    public void SetReference(string refName, string hash)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.GitRepositorySetReference(_handle, refName, hash));
    }
`)
	case "Remote":
		b.WriteString(`
    /// <summary>
    /// Creates a standalone remote for <paramref name="url"/> using in-memory storage.
    /// No local clone is required. Use <see cref="List"/> to enumerate remote references.
    /// </summary>
    public static Remote Create(string url)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitNewRemote(url, out var handle));
        return new Remote(handle);
    }

    /// <summary>The name of this remote as recorded in the repository configuration.</summary>
    public string Name
    {
        get
        {
            ObjectDisposedException.ThrowIf(_disposed, this);
            NativeMethods.ThrowIfError(NativeMethods.GitRemoteConfigName(_handle, out var namePtr));
            return NativeMethods.ConsumeGoString(namePtr)!;
        }
    }

    /// <summary>Returns the full configuration for this remote including URLs and fetch refspecs.</summary>
    public RemoteConfig GetConfig()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.GitRemoteConfig(_handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<RemoteConfig>(json)!;
    }
`)
	case "Submodule":
		b.WriteString(`
    /// <summary>The name of this submodule as recorded in <c>.gitmodules</c>.</summary>
    public string Name
    {
        get
        {
            ObjectDisposedException.ThrowIf(_disposed, this);
            NativeMethods.ThrowIfError(NativeMethods.GitSubmoduleConfigName(_handle, out var namePtr));
            return NativeMethods.ConsumeGoString(namePtr)!;
        }
    }

    /// <summary>Returns the configuration for this submodule (name, path, URL, branch).</summary>
    public SubmoduleConfig GetConfig()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.GitSubmoduleConfig(_handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<SubmoduleConfig>(json)!;
    }

    /// <summary>Returns the current sync status of this submodule.</summary>
    public SubmoduleStatusInfo GetStatus()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.GitSubmoduleStatus(_handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<SubmoduleStatusInfo>(json)!;
    }
`)
	case "Commit":
		for _, field := range []string{"Author", "Committer"} {
			fmt.Fprintf(b, `
    public string %sName
    {
        get
        {
            ObjectDisposedException.ThrowIf(_disposed, this);
            NativeMethods.ThrowIfError(NativeMethods.GitCommit%sName(_handle, out var ptr));
            return NativeMethods.ConsumeGoString(ptr)!;
        }
    }

    public string %sEmail
    {
        get
        {
            ObjectDisposedException.ThrowIf(_disposed, this);
            NativeMethods.ThrowIfError(NativeMethods.GitCommit%sEmail(_handle, out var ptr));
            return NativeMethods.ConsumeGoString(ptr)!;
        }
    }

    public DateTimeOffset %sWhen
    {
        get
        {
            ObjectDisposedException.ThrowIf(_disposed, this);
            NativeMethods.ThrowIfError(NativeMethods.GitCommit%sWhen(_handle, out var ts));
            return DateTimeOffset.FromUnixTimeSeconds(ts);
        }
    }
`, field, field, field, field, field, field)
		}
	}
}
