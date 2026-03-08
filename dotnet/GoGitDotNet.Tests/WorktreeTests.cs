using Xunit;
using File = System.IO.File;

namespace GoGitDotNet.Tests;

public class WorktreeTests : IDisposable
{
    private readonly string _tmpDir;
    private readonly Repository _repo;
    private readonly Worktree _worktree;

    public WorktreeTests()
    {
        _tmpDir = Path.Combine(Path.GetTempPath(), "gogitdotnet-" + Path.GetRandomFileName());
        Directory.CreateDirectory(_tmpDir);
        _repo = Repository.Init(_tmpDir);
        _worktree = _repo.Worktree();
    }

    public void Dispose()
    {
        _worktree.Dispose();
        _repo.Dispose();
        Directory.Delete(_tmpDir, recursive: true);
        GC.SuppressFinalize(this);
    }

    private CommitOptions AuthorOpts() =>
        new CommitOptions().SetAuthor("Alice", "alice@example.com").SetCommitter("Alice", "alice@example.com");

    // --- Status ---

    [Fact]
    public void Status_FreshRepo_IsEmpty()
    {
        var status = _worktree.Status();
        Assert.Empty(status);
    }

    [Fact]
    public void Status_AfterWritingFile_ShowsUntracked()
    {
        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "hello.txt"), "hello\n");
        var status = _worktree.Status();
        Assert.True(status.ContainsKey("hello.txt"));
        Assert.Equal("?", status["hello.txt"].Worktree);
    }

    [Fact]
    public void Status_AfterAdd_ShowsStagedAdded()
    {
        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "hello.txt"), "hello\n");
        _worktree.Add("hello.txt");
        var status = _worktree.Status();
        Assert.True(status.ContainsKey("hello.txt"));
        Assert.Equal("A", status["hello.txt"].Staging);
    }

    // --- Commit ---

    [Fact]
    public void Commit_SingleFile_ReturnsHash()
    {
        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "file.txt"), "content\n");
        _worktree.Add("file.txt");
        using var opts = AuthorOpts();
        var hash = _worktree.Commit("Add file.txt\n", opts);
        Assert.Matches("^[0-9a-f]{40}$", hash);
    }

    [Fact]
    public void Status_AfterCommit_IsClean()
    {
        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "file.txt"), "content\n");
        _worktree.Add("file.txt");
        using var opts = AuthorOpts();
        _worktree.Commit("Add file.txt\n", opts);
        var status = _worktree.Status();
        Assert.Empty(status);
    }

    // --- Read back via object store ---

    [Fact]
    public void Commit_BlobReadableViaObjectStore()
    {
        var content = "round-trip content\n";
        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "data.txt"), content);
        _worktree.Add("data.txt");
        using var opts = AuthorOpts();
        var commitHash = _worktree.Commit("Add data.txt\n", opts);

        using var commit = _repo.CommitObject(commitHash);
        using var tree = commit.Tree();
        var entries = _repo.TreeEntries(tree.Hash);
        var entry = entries.Single(e => e.Name == "data.txt");

        using var blob = _repo.BlobObject(entry.Hash);
        Assert.Equal(content, blob.Contents());
    }

    [Fact]
    public void Commit_MetadataCorrect()
    {
        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "file.txt"), "x\n");
        _worktree.Add("file.txt");
        using var opts = AuthorOpts();
        var commitHash = _worktree.Commit("my message\n", opts);

        using var commit = _repo.CommitObject(commitHash);
        Assert.Equal("my message\n", commit.Message);
        Assert.Equal("Alice", commit.AuthorName);
        Assert.Equal("alice@example.com", commit.AuthorEmail);
    }

    [Fact]
    public void Commit_RootCommit_HasNoParents()
    {
        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "file.txt"), "x\n");
        _worktree.Add("file.txt");
        using var opts = AuthorOpts();
        var commitHash = _worktree.Commit("root\n", opts);

        using var commit = _repo.CommitObject(commitHash);
        Assert.Equal(0, commit.NumParents());
    }

    // --- Multi-commit parent chain ---

    [Fact]
    public void SecondCommit_RecordsFirstAsParent()
    {
        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "a.txt"), "v1\n");
        _worktree.Add("a.txt");
        using var opts1 = AuthorOpts();
        var first = _worktree.Commit("first\n", opts1);

        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "a.txt"), "v2\n");
        _worktree.Add("a.txt");
        using var opts2 = AuthorOpts();
        var second = _worktree.Commit("second\n", opts2);

        using var commit = _repo.CommitObject(second);
        Assert.Equal(1, commit.NumParents());
        using var parent = commit.Parent(0);
        Assert.Equal(first, parent.Hash);
    }

    [Fact]
    public void SecondCommit_UpdatedBlobReadableViaObjectStore()
    {
        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "a.txt"), "v1\n");
        _worktree.Add("a.txt");
        using var opts1 = AuthorOpts();
        _worktree.Commit("first\n", opts1);

        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "a.txt"), "v2\n");
        _worktree.Add("a.txt");
        using var opts2 = AuthorOpts();
        var second = _worktree.Commit("second\n", opts2);

        using var commit = _repo.CommitObject(second);
        using var tree = commit.Tree();
        var entries = _repo.TreeEntries(tree.Hash);
        using var blob = _repo.BlobObject(entries.Single(e => e.Name == "a.txt").Hash);
        Assert.Equal("v2\n", blob.Contents());
    }

    // --- Multi-file commit ---

    [Fact]
    public void Commit_MultipleFiles_AllReadableViaObjectStore()
    {
        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "foo.txt"), "foo\n");
        System.IO.File.WriteAllText(Path.Combine(_tmpDir, "bar.txt"), "bar\n");
        _worktree.Add("foo.txt");
        _worktree.Add("bar.txt");
        using var opts = AuthorOpts();
        var commitHash = _worktree.Commit("Add two files\n", opts);

        using var commit = _repo.CommitObject(commitHash);
        using var tree = commit.Tree();
        var entries = _repo.TreeEntries(tree.Hash);

        Assert.Equal(2, entries.Length);

        using var fooBlob = _repo.BlobObject(entries.Single(e => e.Name == "foo.txt").Hash);
        Assert.Equal("foo\n", fooBlob.Contents());

        using var barBlob = _repo.BlobObject(entries.Single(e => e.Name == "bar.txt").Hash);
        Assert.Equal("bar\n", barBlob.Contents());
    }

    // --- WriteFile ---

    [Fact]
    public void WriteFile_String_StagesAndReturnsHash()
    {
        var hash = _worktree.WriteFile("hello.txt", "hello\n");
        Assert.Matches("^[0-9a-f]{40}$", hash);
        var status = _worktree.Status();
        Assert.Equal("A", status["hello.txt"].Staging);
    }

    [Fact]
    public void WriteFile_Bytes_StagesAndReturnsHash()
    {
        var content = System.Text.Encoding.UTF8.GetBytes("binary-ish\n");
        var hash = _worktree.WriteFile("data.bin", content);
        Assert.Matches("^[0-9a-f]{40}$", hash);
        var status = _worktree.Status();
        Assert.Equal("A", status["data.bin"].Staging);
    }

    [Fact]
    public void WriteFile_ThenCommit_ContentReadableViaObjectStore()
    {
        var content = "written via WriteFile\n";
        _worktree.WriteFile("out.txt", content);
        using var opts = AuthorOpts();
        var commitHash = _worktree.Commit("Add out.txt\n", opts);

        using var commit = _repo.CommitObject(commitHash);
        using var file = commit.File("out.txt");
        Assert.Equal(content, file.Contents());
    }

    [Fact]
    public void WriteFile_StringAndBytes_ProduceSameHash()
    {
        var text = "same content\n";
        var h1 = _worktree.WriteFile("a.txt", text);
        var h2 = _worktree.WriteFile("b.txt", System.Text.Encoding.UTF8.GetBytes(text));
        Assert.Equal(h1, h2);
    }

    [Fact]
    public void WriteFile_OverwriteExisting_UpdatesContent()
    {
        _worktree.WriteFile("file.txt", "v1\n");
        using var opts1 = AuthorOpts();
        _worktree.Commit("v1\n", opts1);

        _worktree.WriteFile("file.txt", "v2\n");
        using var opts2 = AuthorOpts();
        var commitHash = _worktree.Commit("v2\n", opts2);

        using var commit = _repo.CommitObject(commitHash);
        using var file = commit.File("file.txt");
        Assert.Equal("v2\n", file.Contents());
    }
}
