package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateCSharpAuth(dir string) error {
	content := csGenHeader + `#nullable enable
using System.Runtime.InteropServices;

namespace GoGit.Interop;

public sealed class Auth : IDisposable
{
    private long _handle;
    private bool _disposed;

    internal Auth(long handle) => _handle = handle;
    internal long Handle => _handle;

    public static Auth BasicHTTP(string username, string password)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitAuthNewBasicHTTP(username, password, out var handle));
        return new Auth(handle);
    }

    public static Auth TokenHTTP(string token)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitAuthNewTokenHTTP(token, out var handle));
        return new Auth(handle);
    }

    public static Auth SSHKeyFromFile(string user, string pemFile, string password = "")
    {
        NativeMethods.ThrowIfError(NativeMethods.GitAuthNewSSHKeyFromFile(user, pemFile, password, out var handle));
        return new Auth(handle);
    }

    public static Auth SSHKey(string user, string pem, string password = "")
    {
        NativeMethods.ThrowIfError(NativeMethods.GitAuthNewSSHKey(user, pem, password, out var handle));
        return new Auth(handle);
    }

    public static Auth SSHAgent(string user)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitAuthNewSSHAgent(user, out var handle));
        return new Auth(handle);
    }

    public static Auth SSHPassword(string user, string password)
    {
        NativeMethods.ThrowIfError(NativeMethods.GitAuthNewSSHPassword(user, password, out var handle));
        return new Auth(handle);
    }

    public Auth SetInsecureIgnoreHostKey()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.GitAuthSetInsecureIgnoreHostKey(_handle));
        return this;
    }

    public Auth SetKnownHostsFiles(params string[] paths)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        var json = System.Text.Json.JsonSerializer.Serialize(paths);
        NativeMethods.ThrowIfError(NativeMethods.GitAuthSetKnownHostsFiles(_handle, json));
        return this;
    }

    public delegate string? GitHostKeyCallback(string hostname, string remoteAddr, string keyType, string keyBase64);

    private GCHandle? _hostKeyCallbackHandle;

    public Auth SetHostKeyCallback(GitHostKeyCallback callback)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);

        NativeHostKeyCallback native = (hostname, remoteAddr, keyType, keyBase64, _) =>
        {
            var result = callback(
                Marshal.PtrToStringUTF8(hostname)!,
                Marshal.PtrToStringUTF8(remoteAddr)!,
                Marshal.PtrToStringUTF8(keyType)!,
                Marshal.PtrToStringUTF8(keyBase64)!);
            return result != null ? Marshal.StringToCoTaskMemUTF8(result) : IntPtr.Zero;
        };

        _hostKeyCallbackHandle?.Free();
        _hostKeyCallbackHandle = GCHandle.Alloc(native);
        var fnPtr = Marshal.GetFunctionPointerForDelegate(native);
        NativeMethods.ThrowIfError(NativeMethods.GitAuthSetHostKeyCallback(_handle, fnPtr, IntPtr.Zero));
        return this;
    }

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr NativeHostKeyCallback(IntPtr hostname, IntPtr remoteAddr, IntPtr keyType, IntPtr keyBase64, IntPtr userData);

    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;
        _hostKeyCallbackHandle?.Free();
        _hostKeyCallbackHandle = null;
        NativeMethods.GitAuthFree(_handle);
        _handle = 0;
    }
}
`
	return os.WriteFile(filepath.Join(dir, "Auth.cs"), []byte(content), 0644)
}

func generateCSharpSigner(dir string) error {
	content := csGenHeader + `#nullable enable
namespace GoGit.Interop;

public sealed class Signer : IDisposable
{
    private long _handle;
    private bool _disposed;

    internal Signer(long handle) => _handle = handle;
    internal long Handle => _handle;

    public static Signer FromPGPKey(string armoredKey, string passphrase = "")
    {
        NativeMethods.ThrowIfError(NativeMethods.GitSignerNewPGP(armoredKey, passphrase, out var handle));
        return new Signer(handle);
    }

    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;
        NativeMethods.GitSignerFree(_handle);
        _handle = 0;
    }
}
`
	return os.WriteFile(filepath.Join(dir, "Signer.cs"), []byte(content), 0644)
}

func generateCSharpIterators(dir string) error {
	iterDir := filepath.Join(dir, "Iterators")

	commitIter := csGenHeader + `#nullable enable
namespace GoGit.Interop;

public sealed class CommitIterator : IEnumerable<Commit>, IDisposable
{
    private long _handle;
    private bool _disposed;

    internal CommitIterator(long handle) => _handle = handle;

    public IEnumerator<Commit> GetEnumerator()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        while (true)
        {
            var err = NativeMethods.GitCommitIterNext(
                _handle,
                out var commitHandle, out var eof);
            NativeMethods.ThrowIfError(err);

            if (eof != 0) yield break;

            yield return new Commit(commitHandle);
        }
    }

    System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => GetEnumerator();

    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;
        NativeMethods.GitCommitIterFree(_handle);
        _handle = 0;
    }
}
`
	if err := os.WriteFile(filepath.Join(iterDir, "CommitIterator.cs"), []byte(commitIter), 0644); err != nil {
		return err
	}

	refIter := csGenHeader + `#nullable enable
namespace GoGit.Interop;

public sealed class ReferenceIterator : IEnumerable<ReferenceInfo>, IDisposable
{
    private long _handle;
    private bool _disposed;

    internal ReferenceIterator(long handle) => _handle = handle;

    public IEnumerator<ReferenceInfo> GetEnumerator()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        while (true)
        {
            var err = NativeMethods.GitReferenceIterNext(
                _handle,
                out var refName, out var hash, out var eof);
            NativeMethods.ThrowIfError(err);

            if (eof != 0) yield break;

            yield return new ReferenceInfo
            {
                Name = NativeMethods.ConsumeGoString(refName)!,
                Hash = NativeMethods.ConsumeGoString(hash)!,
            };
        }
    }

    System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => GetEnumerator();

    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;
        NativeMethods.GitReferenceIterFree(_handle);
        _handle = 0;
    }
}
`
	if err := os.WriteFile(filepath.Join(iterDir, "ReferenceIterator.cs"), []byte(refIter), 0644); err != nil {
		return err
	}

	fileIter := csGenHeader + `#nullable enable
namespace GoGit.Interop;

public sealed class FileIterator : IEnumerable<File>, IDisposable
{
    private long _handle;
    private bool _disposed;

    internal FileIterator(long handle) => _handle = handle;

    public IEnumerator<File> GetEnumerator()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        while (true)
        {
            var err = NativeMethods.GitFileIterNext(
                _handle,
                out var fileHandle, out var eof);
            NativeMethods.ThrowIfError(err);

            if (eof != 0) yield break;

            yield return new File(fileHandle);
        }
    }

    System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => GetEnumerator();

    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;
        NativeMethods.GitFileIterFree(_handle);
        _handle = 0;
    }
}
`
	if err := os.WriteFile(filepath.Join(iterDir, "FileIterator.cs"), []byte(fileIter), 0644); err != nil {
		return err
	}

	treeIter := csGenHeader + `#nullable enable
namespace GoGit.Interop;

public sealed class TreeIterator : IEnumerable<Tree>, IDisposable
{
    private long _handle;
    private bool _disposed;

    internal TreeIterator(long handle) => _handle = handle;

    public IEnumerator<Tree> GetEnumerator()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        while (true)
        {
            var err = NativeMethods.GitTreeIterNext(
                _handle,
                out var treeHandle, out var eof);
            NativeMethods.ThrowIfError(err);

            if (eof != 0) yield break;

            yield return new Tree(treeHandle);
        }
    }

    System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => GetEnumerator();

    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;
        NativeMethods.GitTreeIterFree(_handle);
        _handle = 0;
    }
}
`
	if err := os.WriteFile(filepath.Join(iterDir, "TreeIterator.cs"), []byte(treeIter), 0644); err != nil {
		return err
	}

	blobIter := csGenHeader + `#nullable enable
namespace GoGit.Interop;

public sealed class BlobIterator : IEnumerable<Blob>, IDisposable
{
    private long _handle;
    private bool _disposed;

    internal BlobIterator(long handle) => _handle = handle;

    public IEnumerator<Blob> GetEnumerator()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        while (true)
        {
            var err = NativeMethods.GitBlobIterNext(
                _handle,
                out var blobHandle, out var eof);
            NativeMethods.ThrowIfError(err);

            if (eof != 0) yield break;

            yield return new Blob(blobHandle);
        }
    }

    System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => GetEnumerator();

    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;
        NativeMethods.GitBlobIterFree(_handle);
        _handle = 0;
    }
}
`
	if err := os.WriteFile(filepath.Join(iterDir, "BlobIterator.cs"), []byte(blobIter), 0644); err != nil {
		return err
	}

	tagIter := csGenHeader + `#nullable enable
namespace GoGit.Interop;

public sealed class TagIterator : IEnumerable<Tag>, IDisposable
{
    private long _handle;
    private bool _disposed;

    internal TagIterator(long handle) => _handle = handle;

    public IEnumerator<Tag> GetEnumerator()
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        while (true)
        {
            var err = NativeMethods.GitTagIterNext(
                _handle,
                out var tagHandle, out var eof);
            NativeMethods.ThrowIfError(err);

            if (eof != 0) yield break;

            yield return new Tag(tagHandle);
        }
    }

    System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => GetEnumerator();

    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;
        NativeMethods.GitTagIterFree(_handle);
        _handle = 0;
    }
}
`
	return os.WriteFile(filepath.Join(iterDir, "TagIterator.cs"), []byte(tagIter), 0644)
}

func generateCSharpModels(dir string) error {
	modelsDir := filepath.Join(dir, "Models")

	// Clean up stale model files that are now handle types
	os.Remove(filepath.Join(modelsDir, "Commit.cs"))

	refInfo := csGenHeader + `#nullable enable
using System.Text.Json.Serialization;

namespace GoGit.Interop;

public sealed class ReferenceInfo
{
    [JsonPropertyName("name")]
    public required string Name { get; init; }

    [JsonPropertyName("hash")]
    public required string Hash { get; init; }
}
`
	if err := os.WriteFile(filepath.Join(modelsDir, "ReferenceInfo.cs"), []byte(refInfo), 0644); err != nil {
		return err
	}

	fileStatus := csGenHeader + `#nullable enable
using System.Text.Json.Serialization;

namespace GoGit.Interop;

public sealed class FileStatus
{
    [JsonPropertyName("staging")]
    public string Staging { get; init; } = "";

    [JsonPropertyName("worktree")]
    public string Worktree { get; init; } = "";

    [JsonPropertyName("extra")]
    public string? Extra { get; init; }
}
`
	if err := os.WriteFile(filepath.Join(modelsDir, "FileStatus.cs"), []byte(fileStatus), 0644); err != nil {
		return err
	}

	goGitException := csGenHeader + `#nullable enable
namespace GoGit.Interop;

public sealed class GoGitException : Exception
{
    public GoGitException(string message) : base(message) { }
}
`
	if err := os.WriteFile(filepath.Join(dir, "GoGitException.cs"), []byte(goGitException), 0644); err != nil {
		return err
	}

	branchConfig := csGenHeader + `#nullable enable
using System.Text.Json.Serialization;

namespace GoGit.Interop;

public sealed class BranchConfig
{
    [JsonPropertyName("name")]
    public string Name { get; init; } = "";

    [JsonPropertyName("remote")]
    public string Remote { get; init; } = "";

    [JsonPropertyName("merge")]
    public string Merge { get; init; } = "";

    [JsonPropertyName("rebase")]
    public string Rebase { get; init; } = "";

    [JsonPropertyName("description")]
    public string? Description { get; init; }
}
`
	if err := os.WriteFile(filepath.Join(modelsDir, "BranchConfig.cs"), []byte(branchConfig), 0644); err != nil {
		return err
	}

	remoteConfig := csGenHeader + `#nullable enable
using System.Text.Json.Serialization;

namespace GoGit.Interop;

public sealed class RemoteConfig
{
    [JsonPropertyName("name")]
    public string Name { get; init; } = "";

    [JsonPropertyName("urls")]
    public string[] URLs { get; init; } = [];

    [JsonPropertyName("fetch")]
    public string[] Fetch { get; init; } = [];
}
`
	if err := os.WriteFile(filepath.Join(modelsDir, "RemoteConfig.cs"), []byte(remoteConfig), 0644); err != nil {
		return err
	}

	submoduleConfig := csGenHeader + `#nullable enable
using System.Text.Json.Serialization;

namespace GoGit.Interop;

public sealed class SubmoduleConfig
{
    [JsonPropertyName("name")]
    public string Name { get; init; } = "";

    [JsonPropertyName("path")]
    public string Path { get; init; } = "";

    [JsonPropertyName("url")]
    public string URL { get; init; } = "";

    [JsonPropertyName("branch")]
    public string Branch { get; init; } = "";
}
`
	if err := os.WriteFile(filepath.Join(modelsDir, "SubmoduleConfig.cs"), []byte(submoduleConfig), 0644); err != nil {
		return err
	}

	submoduleStatusInfo := csGenHeader + `#nullable enable
using System.Text.Json.Serialization;

namespace GoGit.Interop;

public sealed class SubmoduleStatusInfo
{
    [JsonPropertyName("path")]
    public string Path { get; init; } = "";

    [JsonPropertyName("current")]
    public string Current { get; init; } = "";

    [JsonPropertyName("expected")]
    public string Expected { get; init; } = "";

    [JsonPropertyName("branch")]
    public string Branch { get; init; } = "";
}
`
	if err := os.WriteFile(filepath.Join(modelsDir, "SubmoduleStatusInfo.cs"), []byte(submoduleStatusInfo), 0644); err != nil {
		return err
	}

	fileStat := csGenHeader + `#nullable enable
using System.Text.Json.Serialization;

namespace GoGit.Interop;

public sealed class FileStat
{
    [JsonPropertyName("name")]
    public string Name { get; init; } = "";

    [JsonPropertyName("addition")]
    public int Addition { get; init; }

    [JsonPropertyName("deletion")]
    public int Deletion { get; init; }
}
`
	if err := os.WriteFile(filepath.Join(modelsDir, "FileStat.cs"), []byte(fileStat), 0644); err != nil {
		return err
	}

	diffChange := csGenHeader + `#nullable enable
using System.Text.Json.Serialization;

namespace GoGit.Interop;

public sealed class DiffChange
{
    [JsonPropertyName("action")]
    public string Action { get; init; } = "";

    [JsonPropertyName("fromPath")]
    public string? FromPath { get; init; }

    [JsonPropertyName("toPath")]
    public string? ToPath { get; init; }

    [JsonPropertyName("fromHash")]
    public string? FromHash { get; init; }

    [JsonPropertyName("toHash")]
    public string? ToHash { get; init; }
}
`
	if err := os.WriteFile(filepath.Join(modelsDir, "DiffChange.cs"), []byte(diffChange), 0644); err != nil {
		return err
	}

	treeEntryInfo := csGenHeader + `#nullable enable
using System.Text.Json.Serialization;

namespace GoGit.Interop;

public sealed class TreeEntryInfo
{
    [JsonPropertyName("name")]
    public string Name { get; init; } = "";

    [JsonPropertyName("hash")]
    public string Hash { get; init; } = "";

    [JsonPropertyName("mode")]
    public uint Mode { get; init; }
}
`
	if err := os.WriteFile(filepath.Join(modelsDir, "TreeEntryInfo.cs"), []byte(treeEntryInfo), 0644); err != nil {
		return err
	}

	blameResult := csGenHeader + `#nullable enable
using System.Text.Json.Serialization;

namespace GoGit.Interop;

public sealed class BlameResult
{
    [JsonPropertyName("path")]
    public string Path { get; init; } = "";

    [JsonPropertyName("rev")]
    public string Rev { get; init; } = "";

    [JsonPropertyName("lines")]
    public BlameLine[] Lines { get; init; } = [];
}

public sealed class BlameLine
{
    [JsonPropertyName("author")]
    public string Author { get; init; } = "";

    [JsonPropertyName("authorEmail")]
    public string AuthorEmail { get; init; } = "";

    [JsonPropertyName("hash")]
    public string Hash { get; init; } = "";

    [JsonPropertyName("date")]
    public long Date { get; init; }

    [JsonPropertyName("text")]
    public string Text { get; init; } = "";

    public DateTimeOffset DateTimeOffset => DateTimeOffset.FromUnixTimeSeconds(Date);
}
`
	if err := os.WriteFile(filepath.Join(modelsDir, "BlameResult.cs"), []byte(blameResult), 0644); err != nil {
		return err
	}

	gitConfig := csGenHeader + `#nullable enable
using System.Text.Json.Serialization;

namespace GoGit.Interop;

public sealed class GitConfig
{
    [JsonPropertyName("core")]
    public GitConfigCore Core { get; init; } = new();

    [JsonPropertyName("user")]
    public GitConfigIdentity User { get; init; } = new();

    [JsonPropertyName("author")]
    public GitConfigIdentity Author { get; init; } = new();

    [JsonPropertyName("committer")]
    public GitConfigIdentity Committer { get; init; } = new();

    [JsonPropertyName("init")]
    public GitConfigInit Init { get; init; } = new();
}

public sealed class GitConfigCore
{
    [JsonPropertyName("isBare")]
    public bool IsBare { get; init; }

    [JsonPropertyName("worktree")]
    public string Worktree { get; init; } = "";
}

public sealed class GitConfigIdentity
{
    [JsonPropertyName("name")]
    public string Name { get; init; } = "";

    [JsonPropertyName("email")]
    public string Email { get; init; } = "";
}

public sealed class GitConfigInit
{
    [JsonPropertyName("defaultBranch")]
    public string DefaultBranch { get; init; } = "";
}
`
	return os.WriteFile(filepath.Join(modelsDir, "GitConfig.cs"), []byte(gitConfig), 0644)
}

func nativeMethodsAuthSection(b *strings.Builder) {
	b.WriteString(`    // Auth constructors

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitAuthNewBasicHTTP(
        [MarshalAs(UnmanagedType.LPUTF8Str)] string username,
        [MarshalAs(UnmanagedType.LPUTF8Str)] string password,
        out long handleOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitAuthNewTokenHTTP(
        [MarshalAs(UnmanagedType.LPUTF8Str)] string token,
        out long handleOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitAuthNewSSHKeyFromFile(
        [MarshalAs(UnmanagedType.LPUTF8Str)] string user,
        [MarshalAs(UnmanagedType.LPUTF8Str)] string pemFile,
        [MarshalAs(UnmanagedType.LPUTF8Str)] string password,
        out long handleOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitAuthNewSSHKey(
        [MarshalAs(UnmanagedType.LPUTF8Str)] string user,
        [MarshalAs(UnmanagedType.LPUTF8Str)] string pem,
        [MarshalAs(UnmanagedType.LPUTF8Str)] string password,
        out long handleOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitAuthNewSSHAgent(
        [MarshalAs(UnmanagedType.LPUTF8Str)] string user,
        out long handleOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitAuthNewSSHPassword(
        [MarshalAs(UnmanagedType.LPUTF8Str)] string user,
        [MarshalAs(UnmanagedType.LPUTF8Str)] string password,
        out long handleOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern void GitAuthFree(long handle);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitAuthSetInsecureIgnoreHostKey(long handle);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitAuthSetKnownHostsFiles(
        long handle,
        [MarshalAs(UnmanagedType.LPUTF8Str)] string filesJson);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitAuthSetHostKeyCallback(long handle, IntPtr fn, IntPtr userData);

    // Signer constructors

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitSignerNewPGP(
        [MarshalAs(UnmanagedType.LPUTF8Str)] string armoredKey,
        [MarshalAs(UnmanagedType.LPUTF8Str)] string passphrase,
        out long handleOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern void GitSignerFree(long handle);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern void GitSigningKeyFree(long handle);

`)
}

func nativeMethodsIteratorSection(b *strings.Builder) {
	b.WriteString(`    // Iterators

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitCommitIterNext(
        long iterHandle,
        out long handleOut,
        out int eofOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern void GitCommitIterFree(long iterHandle);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitReferenceIterNext(
        long iterHandle,
        out IntPtr refNameOut,
        out IntPtr hashOut,
        out int eofOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern void GitReferenceIterFree(long iterHandle);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitFileIterNext(
        long iterHandle,
        out long handleOut,
        out int eofOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern void GitFileIterFree(long iterHandle);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitTreeIterNext(
        long iterHandle,
        out long handleOut,
        out int eofOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern void GitTreeIterFree(long iterHandle);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitBlobIterNext(
        long iterHandle,
        out long handleOut,
        out int eofOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern void GitBlobIterFree(long iterHandle);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr GitTagIterNext(
        long iterHandle,
        out long handleOut,
        out int eofOut);

    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    public static extern void GitTagIterFree(long iterHandle);

`)
}

func nativeMethodsHandleSection(b *strings.Builder, pkg *Package) {
	for _, ht := range pkg.Types {
		fmt.Fprintf(b, "    // %s methods\n\n", ht.GoName)
		for _, m := range ht.Methods {
			if isOverrideMethod(ht.GoName, m.GoName) {
				writeOverrideNativeMethod(b, &ht, m)
			} else {
				writeGenericNativeMethod(b, &ht, m)
			}
		}
		writeExtraNativeMethods(b, &ht)
		writeFieldGetterNativeMethods(b, &ht)
		writeFreeDllImport(b, &ht)
	}
}

func writeFieldGetterNativeMethods(b *strings.Builder, ht *HandleType) {
	for _, f := range ht.Fields {
		handleParam := strings.ToLower(ht.GoName[:1]) + "Handle"
		var outType string
		switch f.Mapping.Kind {
		case MappingString, MappingHash, MappingReferenceName:
			outType = "out IntPtr"
		case MappingPrimitive:
			switch f.Mapping.CSharpType {
			case "long":
				outType = "out long"
			case "uint":
				outType = "out uint"
			default:
				outType = "out int"
			}
		case MappingBool:
			outType = "out int"
		default:
			continue
		}

		b.WriteString("    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
		fmt.Fprintf(b, "    public static extern IntPtr %s(long %s, %s valOut);\n\n", f.CGetterName, handleParam, outType)
	}
}

func writeGenericNativeMethod(b *strings.Builder, ht *HandleType, m Method) {
	handleParam := strings.ToLower(ht.GoName[:1]) + "Handle"
	var params []string
	params = append(params, fmt.Sprintf("long %s", handleParam))

	for _, p := range m.Params {
		params = append(params, csNativeParam(p))
	}

	for _, r := range m.Returns {
		if r.IsError {
			continue
		}
		switch r.Mapping.Kind {
		case MappingReference:
			params = append(params, "out IntPtr refNameOut", "out IntPtr hashOut")
		case MappingHandle, MappingIterator:
			params = append(params, "out long handleOut")
		case MappingString, MappingReferenceName, MappingHash, MappingRevision, MappingStringSlice:
			params = append(params, "out IntPtr strOut")
		case MappingBool:
			params = append(params, "out int boolOut")
		case MappingEnum:
			params = append(params, "out int valOut")
		case MappingPrimitive:
			if r.CType == "C.longlong" {
				params = append(params, "out long valOut")
			} else {
				params = append(params, "out int valOut")
			}
		case MappingTime, MappingDuration:
			params = append(params, "out long valOut")
		}
	}

	writeDllImport(b, m.CName, strings.Join(params, ", "))
}

func csSafeName(name string) string {
	switch name {
	case "in", "out", "ref", "string", "object", "event", "base", "params":
		return "@" + name
	}
	return name
}

func csNativeParam(p Param) string {
	name := csSafeName(p.CName)
	switch p.Mapping.Kind {
	case MappingString, MappingReferenceName, MappingHash, MappingRevision, MappingStringSlice:
		return fmt.Sprintf("[MarshalAs(UnmanagedType.LPUTF8Str)] string %s", name)
	case MappingBool, MappingEnum:
		return fmt.Sprintf("int %s", name)
	case MappingPrimitive:
		switch p.Mapping.CSharpType {
		case "long":
			return fmt.Sprintf("long %s", name)
		case "uint":
			return fmt.Sprintf("uint %s", name)
		default:
			return fmt.Sprintf("int %s", name)
		}
	case MappingTime, MappingDuration:
		return fmt.Sprintf("long %s", name)
	case MappingOptions, MappingHandle, MappingAuth, MappingSigner:
		return fmt.Sprintf("long %s", name)
	default:
		return fmt.Sprintf("long %s", name)
	}
}

func writeDllImport(b *strings.Builder, name, params string) {
	b.WriteString("    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
	fmt.Fprintf(b, "    public static extern IntPtr %s(\n        %s);\n\n", name, params)
}

func writeFreeDllImport(b *strings.Builder, ht *HandleType) {
	handleParam := strings.ToLower(ht.GoName[:1]) + "Handle"
	b.WriteString("    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
	fmt.Fprintf(b, "    public static extern void %sFree(long %s);\n\n", ht.CPrefix, handleParam)
}
