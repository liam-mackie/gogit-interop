# gogit-interop

A code generator that produces a C-shared library wrapping [go-git v6](https://github.com/go-git/go-git) and matching C# P/Invoke bindings, enabling .NET applications to use go-git's pure-Go git implementation without requiring libgit2 or shelling out to the `git` CLI.

## Why this exists

go-git is the most complete pure-Go implementation of git. It supports cloning, committing, pushing, pulling, diffing, blaming, and more — all without requiring a native git installation. However, go-git only has a Go API. This project bridges that gap by:

1. **Reflecting** on go-git's Go types at build time using `go/packages` and `go/types`
2. **Generating** a CGO shared library (`libgogit.dylib`/`.so`/`.dll`) that exposes Go functions via C-compatible exports
3. **Generating** a matching C# class library (`GoGit.Interop`) with P/Invoke declarations and idiomatic wrapper classes

The result is a NuGet package that .NET applications reference like any other library — the native binary is bundled as a runtime-specific asset.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  C# Application                                        │
│    using repo = Repository.Open("/path/to/repo");      │
│    var commit = repo.GetCommitObject(hash);             │
└──────────────────────┬──────────────────────────────────┘
                       │ P/Invoke (DllImport)
┌──────────────────────▼──────────────────────────────────┐
│  GoGit.Interop (C#)                                     │
│    NativeMethods.cs  — DllImport declarations           │
│    Repository.cs     — idiomatic wrapper class          │
│    CloneOptions.cs   — fluent options builder           │
│    CommitIterator.cs — IEnumerable<Commit> wrapper      │
└──────────────────────┬──────────────────────────────────┘
                       │ C ABI (exported functions)
┌──────────────────────▼──────────────────────────────────┐
│  libgogit (Go c-shared)                                 │
│    repository_gen.go — //export GitRepositoryCommitObject│
│    options_gen.go    — //export GitCloneOptionsSetURL    │
│    handle.go         — sync.Map-based handle store      │
└──────────────────────┬──────────────────────────────────┘
                       │ Go function calls
┌──────────────────────▼──────────────────────────────────┐
│  go-git v6                                              │
│    *git.Repository, *object.Commit, etc.                │
└─────────────────────────────────────────────────────────┘
```

### Handle-based FFI

Go objects cannot be passed directly across the FFI boundary — the Go garbage collector could move them, and C/C# has no way to interact with Go pointers. Instead, every Go object that crosses the boundary is stored in a thread-safe `sync.Map` keyed by a monotonically increasing `int64` handle:

```go
// shared/handle.go
var handles sync.Map

func storeHandle(obj any) int64 {
    h := handleCounter.Add(1)
    handles.Store(h, obj)
    return h
}

func loadHandle[T any](h int64) (T, bool) {
    v, ok := handles.Load(h)
    if !ok { var zero T; return zero, false }
    t, ok := v.(T)
    return t, ok
}
```

On the C# side, each wrapper class holds the `int64` handle internally and passes it to every native call. `IDisposable` calls the corresponding `Free` export to remove the handle from the Go-side map:

```csharp
public sealed class Repository : IDisposable
{
    private long _handle;
    internal long Handle => _handle;

    public void Dispose()
    {
        NativeMethods.GitRepositoryFree(_handle);
        _handle = 0;
    }
}
```

### Error convention

Every exported Go function that can fail returns `*C.char`. A `nil` return means success; a non-nil return is a UTF-8 error message allocated with `C.CString`. On the C# side, `NativeMethods.ThrowIfError(IntPtr)` reads the string, frees the Go-allocated memory via `GitFreeString`, and throws a `GoGitException`.

### String ownership

All strings returned from Go via `*C.char` out-parameters are allocated with `C.CString` (which calls `malloc`). The C# side reads the string with `Marshal.PtrToStringUTF8`, then immediately frees it via `GitFreeString`. The helper `NativeMethods.ConsumeGoString(IntPtr)` encapsulates this pattern.

## Project layout

```
generate/              Code generator (go run .)
  analyzer.go          Reflects on go-git packages, discovers types
  model.go             Data model: HandleType, Method, OptionsStruct, etc.
  mappings.go          Go type → FFI type mapping (TypeMapping)
  overrides.go         Hand-coded wrappers for complex methods
  filter.go            Type exclusion lists
  gen_go.go            Emits Go CGO wrapper functions
  gen_go_special.go    Emits auth, signing, iterator Go code
  gen_go_options.go    Emits options constructor/setter Go code
  gen_csharp.go        Emits NativeMethods.cs, orchestrates C# output
  gen_csharp_wrappers.go  Emits wrapper classes (Repository.cs, etc.)
  gen_csharp_special.go   Emits Auth.cs, Signer.cs, iterator + model classes
  gen_csharp_options.go   Emits options C# classes

shared/                Go c-shared library (build output)
  main.go              Package main, //go:generate directive
  handle.go            Handle store (sync.Map + atomic counter)
  errors.go            toCError() + GitFreeString export
  callbacks.go         Progress/signing callback wrappers
  callbacks.h          C function pointer typedefs
  *_gen.go             Generated: one per handle type + options + iterators

dotnet/GoGit.Interop/  C# class library (build output)
  NativeMethods.cs     All DllImport declarations
  Repository.cs        Wrapper class with factory methods
  Commit.cs, Tree.cs, ... Wrapper classes for handle types
  Options/             Fluent options builders (CloneOptions.cs, etc.)
  Iterators/           IEnumerable wrappers (CommitIterator.cs, etc.)
  Models/              JSON-deserialized models (BlameResult.cs, etc.)
```

## How the generator works

### Phase 1: Analyze

The generator loads six go-git packages using `golang.org/x/tools/go/packages`:

```go
var packagesToLoad = []string{
    "github.com/go-git/go-git/v6",
    "github.com/go-git/go-git/v6/plumbing",
    "github.com/go-git/go-git/v6/plumbing/object",
    "github.com/go-git/go-git/v6/plumbing/storer",
    "github.com/go-git/go-git/v6/plumbing/transport",
    "github.com/go-git/go-git/v6/config",
}
```

It then:

1. **Discovers enums** — any named integer type with associated constants (e.g., `ResetMode`, `SubmoduleRescursivity`)
2. **Discovers options structs** — any exported struct ending in `Options` (e.g., `CloneOptions`, `FetchOptions`), with mappable fields enumerated
3. **Registers seed handle types** — 9 core types that become wrapper classes:
   - `git.Repository`, `git.Worktree`, `git.Remote`, `git.Submodule`
   - `object.Commit`, `object.Tree`, `object.Blob`, `object.Tag`, `object.File`
4. **Discovers top-level functions** — exported functions in the `git` package (e.g., `PlainClone`, `PlainOpen`)
5. **Processes handle types** — for each seed type, enumerates all exported methods and struct fields, mapping each parameter and return type

### Phase 2: Map types

Every Go type encountered is run through `resolveTypeMapping()`, which returns a `TypeMapping` with the FFI type information:

| Go Type | MappingKind | C Type | C# Type | Notes |
|---------|-------------|--------|---------|-------|
| `string` | MappingString | `*C.char` | `string` | UTF-8, marshal via LPUTF8Str |
| `bool` | MappingBool | `C.int` | `int` | 0/1 encoding |
| `int`, `int32` | MappingPrimitive | `C.int` | `int` | |
| `int64` | MappingPrimitive | `C.longlong` | `long` | |
| `plumbing.Hash` | MappingHash | `*C.char` | `string` | Hex string via `.String()` |
| `plumbing.ReferenceName` | MappingReferenceName | `*C.char` | `string` | Named string type |
| `*git.Repository` | MappingHandle | `C.longlong` | `long` | Handle in sync.Map |
| `*git.CloneOptions` | MappingOptions | `C.longlong` | `long` | Handle in sync.Map |
| `transport.AuthMethod` | MappingAuth | `C.longlong` | `long` | Interface, handle store |
| `object.CommitIter` | MappingIterator | `C.longlong` | `long` | Handle, yields via Next/Free |
| `time.Time` | MappingTime | `C.longlong` | `long` | Unix seconds |
| `time.Duration` | MappingDuration | `C.longlong` | `long` | Nanoseconds |
| `[]string` | MappingStringSlice | `*C.char` | `string` | JSON-encoded array |
| `*plumbing.Reference` | MappingReference | `*C.char` | `string` | Expanded to (name, hash) pair |

If a type cannot be mapped (returns `MappingSkip`), the method or field is skipped with a warning.

### Phase 3: Generate

The generator emits:

- **One Go file per handle type** (`repository_gen.go`, `commit_gen.go`, etc.) containing `//export` functions
- **One Go file for all options** (`options_gen.go`) with constructors, setters, and free functions
- **One Go file for iterators** (`iterators_gen.go`) with Next/Free functions
- **Standalone Go files** for auth and signing (`auth_gen.go`, `signing_gen.go`)
- **One C# NativeMethods.cs** with all DllImport declarations
- **One C# wrapper class per handle type** with idiomatic methods
- **One C# class per options struct** with fluent builder pattern
- **C# iterator classes** implementing `IEnumerable<T>`
- **C# model classes** for JSON-deserialized complex return types

## Type mapping challenges

### Why some types can't cross the FFI boundary

The fundamental constraint is that **only primitive C types can cross a CGO export boundary**: integers, floats, and `*C.char` (pointers to C-allocated strings). Any Go type that isn't representable as one of these needs a mapping strategy.

#### Successfully mapped complex types

| Challenge | Strategy |
|-----------|----------|
| **Go structs** (Repository, Commit, etc.) | Store in `sync.Map`, pass `int64` handle |
| **Go interfaces** (`transport.AuthMethod`) | Store concrete implementation, pass handle |
| **Named string types** (`ReferenceName`, `Revision`) | Convert to/from plain string |
| **Hash/ObjectID** (type alias) | Convert to hex string via `.String()`, parse back with `plumbing.NewHash()` |
| **`*plumbing.Reference`** | Expand to two out-params: `(refName, hash)` |
| **`[]string` and `[]NamedString`** | JSON encode/decode as `["a","b"]` |
| **`time.Time`** | Unix timestamp (int64 seconds) |
| **`time.Duration`** | Nanoseconds (int64) |
| **Integer-based enums** | Cast to/from C `int` |
| **Iterator types** (`CommitIter`, `FileIter`, etc.) | Store as handle, expose Next/Free pattern |
| **Options structs** | Store as handle, expose per-field setters |
| **`*object.Signature`** | Special setter: `SetAuthor(name, email)` → constructs Go struct internally |
| **`transport.ProxyOptions`** | Special setter: `SetProxy(url, username, password)` |
| **`[]plumbing.Hash`** | Special setter: JSON array of hex strings |
| **`*ForceWithLease`** | Special setter: `SetForceWithLease(refName, hash)` |
| **`*openpgp.Entity`** | Separate handle (`GitSigningKeyNewPGP`), passed as handle to setter |
| **Complex return types** (blame, stats, diffs, config) | Override method returns JSON, C# deserializes into model class |

#### Intentionally unmapped types

| Type | Why it can't be mapped |
|------|----------------------|
| **`[]*regexp.Regexp`** | Compiled regex objects can't cross FFI. Would need string patterns + Go-side compilation. Blocks `Grep`/`GrepOptions`. |
| **`object.Object`** (interface) | The return type is an interface that could be any of Commit/Tree/Blob/Tag. Would need runtime type-switching + handle tagging. Blocks `Repository.Object()`. |
| **`*object.ObjectIter`** | Same problem — yields `object.Object` interface values. Blocks `Repository.Objects()`. |
| **`PruneHandler`** (function callback) | Go function types can't be directly exported. Would need a C function pointer callback mechanism. Blocks `Repository.Prune()`. |
| **`*config.Config`** (full) | Deeply nested struct with maps of maps. The override `Repository.Config()` returns a flattened JSON subset. `SetConfig` would require constructing the full struct. |
| **`*RepackConfig`** | Undiscovered options struct (not ending in `Options`). Blocks `Repository.RepackObjects()`. |
| **`*TreeEntry`** (as parameter) | `Tree.TreeEntryFile()` takes a `*TreeEntry`. The override `Tree.FindEntry()` returns entry data as JSON instead. |
| **`*Context` variants** | `PatchContext`, `DiffContext`, `StatsContext`, `ListContext` — the non-context versions are exposed instead, using `context.Background()` internally. |
| **`Decode`/`Encode` methods** | Internal serialization — not useful for consumers. |

### The Hash/ObjectID alias problem

In go-git v6, `plumbing.Hash` is a **type alias** for `plumbing.ObjectID`:

```go
type ObjectID [20]byte
type Hash = ObjectID  // type alias, not a named type
```

This means `Hash` appears as `*types.Alias` in the Go type system, not `*types.Named`. The generator must call `types.Unalias()` before type-switching, otherwise the alias falls through to the default case and gets skipped:

```go
func resolveTypeMapping(t types.Type) TypeMapping {
    if alias, ok := t.(*types.Alias); ok {
        return resolveTypeMapping(types.Unalias(alias))
    }
    // ...
}
```

### Override methods vs. auto-generated methods

Most methods are auto-generated from their Go signature. However, some methods have signatures that can't be directly mapped, or produce return types that need special handling. These are **override methods** — hand-coded Go functions in `overrides.go` with corresponding hand-coded C# wrappers.

Override methods are needed when:
- **Parameters are complex structs** that need to be constructed from simpler inputs (e.g., `CreateRemote` takes `*config.RemoteConfig` → override accepts `name, url` strings)
- **Return types are complex** and need to be serialized to JSON (e.g., `Worktree.Status()` returns `Status` map → override marshals to JSON)
- **The method involves an interface** that we handle internally (e.g., `Merge` takes a `plumbing.Reference` → override accepts `refName, hash` strings)

Current override methods:

| Method | Why overridden |
|--------|---------------|
| `Repository.CreateRemote` | Takes `*config.RemoteConfig` → accepts `(name, url)` |
| `Repository.CreateBranch` | Takes `*config.Branch` → accepts `(name, hash)` |
| `Repository.Merge` | Takes `plumbing.Reference` value → accepts `(refName, hash)` |
| `Repository.Branch` | Returns `*config.Branch` → JSON |
| `Repository.Config` | Returns `*config.Config` → flattened JSON subset |
| `Repository.CreateRemoteAnonymous` | Takes `*config.RemoteConfig` → accepts `url` |
| `Worktree.Status` | Returns `Status` (map type) → JSON |
| `Worktree.Submodules` | Returns `Submodules` → JSON array of names |
| `Remote.List` | Returns `[]*plumbing.Reference` → JSON array |
| `Blob.Reader` | Returns `io.ReadCloser` → reads all bytes, returns string |
| `Commit.Stats` | Returns `FileStats` → JSON array |
| `Commit.Patch` | Returns `*Patch` → unified diff string |
| `Commit.MergeBase` | Returns `[]*Commit` → JSON array of hashes |
| `Commit.Verify` | Returns `*openpgp.Entity` → void (ignore entity, just check error) |
| `Tree.Diff` | Returns `Changes` → JSON array |
| `Tree.Patch` | Returns `*Patch` → unified diff string |
| `Tree.FindEntry` | Returns `*TreeEntry` → JSON |
| `Tag.Verify` | Same as Commit.Verify |

### Extra methods

Some useful operations don't correspond to a single Go method, or they access nested struct fields. These are **extra methods** — additional exports not discovered by reflection:

- `Repository.Remotes()` — calls `repo.Remotes()`, returns JSON array of names
- `Repository.Blame(commit, path)` — calls `git.Blame()`, returns JSON
- `Remote.Create(url)` — static factory; creates a standalone remote via `git.NewRemote` with in-memory storage, no local clone required
- `Remote.Config()` — calls `remote.Config()`, returns full JSON
- `Submodule.Config()` / `Submodule.Status()` — accesses nested config/status
- `Commit.AuthorName`, `AuthorEmail`, `AuthorWhen`, etc. — accesses `Signature` struct fields

### Iterator pattern

Go iterators follow the `Next() → (value, error)` pattern where `io.EOF` signals completion. The generator wraps these as handle-based iterators:

**Go side:**
```go
//export GitCommitIterNext
func GitCommitIterNext(iterHandle C.longlong, handleOut *C.longlong, eofOut *C.int) *C.char {
    iter, ok := loadHandle[object.CommitIter](int64(iterHandle))
    // ...call iter.Next(), store result, signal EOF
}
```

**C# side:**
```csharp
public sealed class CommitIterator : IEnumerable<Commit>, IDisposable
{
    public IEnumerator<Commit> GetEnumerator()
    {
        while (true)
        {
            var err = NativeMethods.GitCommitIterNext(_handle, out var commitHandle, out var eof);
            NativeMethods.ThrowIfError(err);
            if (eof != 0) yield break;
            yield return new Commit(commitHandle);
        }
    }
}
```

Six iterator types are supported: `CommitIterator`, `ReferenceIterator`, `FileIterator`, `TreeIterator`, `BlobIterator`, `TagIterator`.

### Options builder pattern

Go options structs are exposed as fluent builders in C#:

```csharp
using var opts = new CloneOptions()
    .SetURL("https://github.com/user/repo.git")
    .SetAuth(Auth.TokenHTTP(token))
    .SetDepth(1)
    .SetSingleBranch(true)
    .SetProxy("http://proxy:8080");

using var repo = Repository.Clone("/tmp/repo", opts);
```

Each option struct gets a `New` constructor (allocates Go struct, returns handle), per-field `Set` functions, and a `Free` function. Special setters handle fields that can't be directly mapped (signatures, proxy options, hash arrays, etc.).

## How to add support for new go-git types

When go-git adds new types or methods, follow this process to determine what needs mapping.

### Step 1: Run the generator and review warnings

```bash
cd generate && go run . 2>&1 | grep WARNING
```

Each warning tells you exactly what was skipped and why:
- `skipping FooOptions.Bar — unmappable type X` → an options field needs a special setter
- `skipping Foo.Method — unmappable signature` → a method has an unmappable parameter or return type

### Step 2: Determine the mapping strategy

Ask these questions about each unmapped type:

1. **Is it a new struct that should be a handle type?** If it's a major type that users create, hold onto, and call methods on, add it to `registerSeedHandleTypes()` in `analyzer.go`. The generator will automatically discover its methods and fields.

2. **Is it a new options struct?** If it ends in `Options` and is a struct, it's auto-discovered. If it has a different naming convention (like `RepackConfig`), you'd need to explicitly register it.

3. **Is it a new iterator type?** Add it to `knownPointerIterators` in `mappings.go`, add Go Next/Free functions in `gen_go_special.go`, add a C# iterator class in `gen_csharp_special.go`, add DllImports in `nativeMethodsIteratorSection`, and add a case to `csIteratorClassName`.

4. **Is it a complex return type?** Add it as an override method in `overrides.go` that marshals to JSON, add a model class in `gen_csharp_special.go`, and wire up the C# wrapper in `gen_csharp_wrappers.go`.

5. **Is it an unmappable options field?** Add a special setter in `generateOptionsSpecialSetters` (Go side) and `nativeMethodsOptionsSection` + options template data (C# side).

6. **Is it truly unmappable?** Some types can't cross FFI at all (`func` types, `chan`, compiled regex, interface returns). Document these in the intentionally excluded list.

### Step 3: Add a new seed handle type (example)

If go-git adds a new type like `*object.Worktree` that you want to wrap:

1. **Add to seed types** in `analyzer.go`:
   ```go
   {"github.com/go-git/go-git/v6/plumbing/object", "NewType"},
   ```

2. **Register the mapping** — this happens automatically via `enqueueHandle()` during analysis.

3. **Run the generator** — it discovers all exported methods and fields on the type, generates Go exports and C# wrappers automatically.

4. **Check warnings** — if any methods were skipped, decide if they need override methods.

5. **Add method name overrides** in `csPublicMethodName()` if any method names would clash with the class name or C# reserved words.

### Step 4: Add an override method (example)

To wrap a method with an unmappable signature like `Foo.ComplexMethod(opts *BarConfig) (*BazResult, error)`:

1. **Register the override** in `overrideMethods` map:
   ```go
   "Foo": {"ComplexMethod": true},
   ```

2. **Implement the Go function** in `overrides.go`:
   ```go
   func generateOverrideFooComplexMethod(b *strings.Builder, cName string) {
       // emit //export, load receiver, call method, marshal result to JSON
   }
   ```

3. **Add the dispatch** in `generateOverrideMethod()`.

4. **Add the native method** in `writeOverrideNativeMethod()`.

5. **Add the C# wrapper** in `generateOverrideWrapperMethod()`.

6. **Add the C# model class** (if returning JSON) in `generateCSharpModels()`.

### Step 5: Verify

```bash
cd generate && go run .                                    # should show only expected warnings
cd ../shared && go build -buildmode=c-shared -o /dev/null . # Go compiles
cd ../dotnet/GoGit.Interop && dotnet build                  # C# compiles
```

## Build

### Prerequisites

- Go 1.25+
- .NET 8 SDK
- Platform-specific C cross-compilers for non-host targets:
  - macOS arm64/amd64: Xcode command line tools
  - Linux amd64: `x86_64-linux-musl-gcc`
  - Windows amd64: `x86_64-w64-mingw32-gcc`

### Commands

```bash
make generate              # Run the code generator
make build-darwin-arm64    # Build native library for macOS ARM
make build-darwin-amd64    # Build native library for macOS Intel
make build-linux-amd64     # Build native library for Linux
make build-windows-amd64   # Build native library for Windows
make dev                   # Generate + build current platform + pack NuGet
make pack                  # Pack NuGet (after building all targets)
make clean                 # Remove build artifacts
```

### Quick development cycle

```bash
make dev
# Produces: dotnet/GoGit.Interop/bin/Release/GoGit.Interop.0.1.1-dev.YYYYMMDDHHMMSS.nupkg
```

## C# API usage

```csharp
using GoGit.Interop;

// Open a repository
using var repo = Repository.Open("/path/to/repo");

// Clone with options
using var opts = new CloneOptions()
    .SetURL("https://github.com/user/repo.git")
    .SetAuth(Auth.TokenHTTP(Environment.GetEnvironmentVariable("GH_TOKEN")!))
    .SetDepth(1);
using var cloned = Repository.Clone("/tmp/clone", opts);

// Read commits
using var log = repo.Log();
foreach (var commit in log)
{
    Console.WriteLine($"{commit.Hash}: {commit.Message}");
    Console.WriteLine($"  by {commit.AuthorName} <{commit.AuthorEmail}>");
    commit.Dispose();
}

// Diff two trees
using var commitA = repo.GetCommitObject(hashA);
using var commitB = repo.GetCommitObject(hashB);
using var treeA = commitA.GetTree();
using var treeB = commitB.GetTree();
var changes = treeA.Diff(treeB);
foreach (var change in changes)
    Console.WriteLine($"{change.Action}: {change.FromPath} → {change.ToPath}");

// Get blame
var blame = Repository.Blame(commit, "README.md");
foreach (var line in blame.Lines)
    Console.WriteLine($"{line.Hash[..7]} {line.Author}: {line.Text}");

// Worktree operations
using var wt = repo.GetWorktree();
var status = wt.Status();
foreach (var (path, fs) in status)
    Console.WriteLine($"{fs.Staging}{fs.Worktree} {path}");

// Branch config
var branch = repo.GetBranch("main");
Console.WriteLine($"Tracks: {branch.Remote}/{branch.Merge}");

// Repository config
var config = repo.GetConfig();
Console.WriteLine($"User: {config.User.Name} <{config.User.Email}>");

// List remote refs without a local clone
using var opts = new ListOptions()
    .SetAuth(Auth.TokenHTTP(Environment.GetEnvironmentVariable("GH_TOKEN")!));
using var remote = Remote.Create("https://github.com/user/repo.git");
var refs = remote.List(opts);
foreach (var r in refs)
    Console.WriteLine($"{r.Hash}  {r.Name}");
```

## Current API coverage

### Handle types (9)

| Type | Methods | Fields | Notes |
|------|---------|--------|-------|
| Repository | ~20 | — | Factory methods, object lookups, branch/tag/remote ops |
| Worktree | ~10 | — | Add, commit, checkout, status, reset, restore |
| Remote | ~4 | — | Fetch, push, list refs; `Remote.Create(url)` for standalone use without a local clone |
| Submodule | ~3 | — | Init, update, config, status |
| Commit | ~10 | Hash, Message, etc. | Tree, parents, stats, patch, merge-base, verify |
| Tree | ~6 | Hash | File lookup, entries, diff, patch, find-entry |
| Blob | ~2 | Hash, Size | Contents (via Reader override) |
| Tag | ~6 | Hash, Name, Message, etc. | Target object access, verify |
| File | ~2 | Name, Hash | Contents |

### Options types (24)

AddOptions, CheckoutOptions, CleanOptions, CloneOptions, CommitOptions, CreateTagOptions, DiffTreeOptions, FetchOptions, GrepOptions, ListOptions, LogLimitOptions, LogOptions, MergeOptions, PlainOpenOptions, ProxyOptions, PruneOptions, PullOptions, PushOptions, ReceivePackOptions, ResetOptions, RestoreOptions, StatusOptions, SubmoduleUpdateOptions, UploadPackOptions

### Iterators (6)

CommitIterator, ReferenceIterator, FileIterator, TreeIterator, BlobIterator, TagIterator

### Model classes (13)

ReferenceInfo, FileStatus, BranchConfig, RemoteConfig, SubmoduleConfig, SubmoduleStatusInfo, FileStat, DiffChange, TreeEntryInfo, BlameResult, BlameLine, GitConfig (with nested types), GoGitException
