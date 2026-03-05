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
	b.WriteString("\nnamespace GoGit.Interop;\n\n")
	fmt.Fprintf(&b, "public sealed class %s : IDisposable\n{\n", ht.GoName)
	b.WriteString("    private long _handle;\n")
	b.WriteString("    private bool _disposed;\n\n")
	fmt.Fprintf(&b, "    internal %s(long handle) => _handle = handle;\n", ht.GoName)

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
	case "Repository", "Worktree", "Remote":
		return true
	}
	return false
}

func generateRepoFactoryMethods(b *strings.Builder, pkg *Package) {
	b.WriteString(`
    public static Repository Init(string path, bool isBare = false)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitPlainInit(path, isBare ? 1 : 0, out var handle));
        return new Repository(handle);
    }

    public static Repository Open(string path)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitPlainOpen(path, out var handle));
        return new Repository(handle);
    }

    public static Repository OpenWithOptions(string path, PlainOpenOptions options)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitPlainOpenWithOptions(path, options.Handle, out var handle));
        return new Repository(handle);
    }

    public static Repository Clone(string path, CloneOptions options)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitPlainClone(path, options.Handle, out var handle));
        return new Repository(handle);
    }
`)
}

func generateGenericWrapperMethod(b *strings.Builder, ht *HandleType, m Method) {
	returnType := csWrapperReturnType(m)
	publicName := csPublicMethodName(ht, m)
	params := csWrapperParams(m)

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
			fmt.Fprintf(b, "        NativeMethods.ThrowIfError(NativeMethods.%s(_handle%s, out var h));\n", m.CName, nativeArgs)
			fmt.Fprintf(b, "        return new %s(h);\n", r.Mapping.HandleType)
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
		case MappingPrimitive:
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
		case MappingPrimitive:
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
	case "Worktree.Submodule":
		return "GetSubmodule"
	case "Submodule.Repository":
		return "GetRepository"
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
	default:
		return handleType + "Iterator"
	}
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
		b.WriteString(`    public Remote CreateRemote(string name, string url)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, name, url, out var handle));
        return new Remote(handle);
    }
`)
	case "Repository.CreateBranch":
		b.WriteString(`    public void CreateBranch(string name, string hash)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, name, hash));
    }
`)
	case "Repository.CommitObject":
		b.WriteString(`    public Commit GetCommitObject(string hash)
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
		b.WriteString(`    public void Merge(string refName, string hash, MergeOptions? options = null)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        var optsHandle = options?.Handle ?? 0;
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, refName, hash, optsHandle));
    }
`)
	case "Worktree.Status":
		b.WriteString(`    public Dictionary<string, FileStatus> Status()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<Dictionary<string, FileStatus>>(json) ?? new();
    }
`)
	case "Worktree.Submodules":
		b.WriteString(`    public string[] Submodules()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<string[]>(json) ?? [];
    }
`)
	case "Remote.List":
		b.WriteString(`    public ReferenceInfo[] List(ListOptions options)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.` + m.CName + `(_handle, options.Handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<ReferenceInfo[]>(json) ?? [];
    }
`)
	}
}

func generateExtraWrapperMethods(b *strings.Builder, ht *HandleType) {
	switch ht.GoName {
	case "Repository":
		b.WriteString(`
    public string[] Remotes()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.GitRepositoryRemotes(_handle, out var jsonPtr));
        var json = NativeMethods.ConsumeGoString(jsonPtr)!;
        return JsonSerializer.Deserialize<string[]>(json) ?? [];
    }
`)
	case "Remote":
		b.WriteString(`
    public string Name
    {
        get
        {
            ObjectDisposedException.ThrowIf(_disposed, this);
            NativeMethods.ThrowIfError(NativeMethods.GitRemoteConfigName(_handle, out var namePtr));
            return NativeMethods.ConsumeGoString(namePtr)!;
        }
    }
`)
	case "Submodule":
		b.WriteString(`
    public string Name
    {
        get
        {
            ObjectDisposedException.ThrowIf(_disposed, this);
            NativeMethods.ThrowIfError(NativeMethods.GitSubmoduleConfigName(_handle, out var namePtr));
            return NativeMethods.ConsumeGoString(namePtr)!;
        }
    }
`)
	}
}
