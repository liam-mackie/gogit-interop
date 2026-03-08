using Xunit;

namespace GoGitDotNet.Tests;

public class RepositoryTests : IDisposable
{
    private readonly string _tmpDir;

    public RepositoryTests()
    {
        _tmpDir = Path.Combine(Path.GetTempPath(), "gogitdotnet-" + Path.GetRandomFileName());
        Directory.CreateDirectory(_tmpDir);
    }

    public void Dispose()
    {
        Directory.Delete(_tmpDir, recursive: true);
        GC.SuppressFinalize(this);
    }

    // --- Init ---

    [Fact]
    public void Init_NonBare_CreatesRepository()
    {
        var path = Path.Combine(_tmpDir, "repo");
        using var repo = Repository.Init(path);
        Assert.True(Directory.Exists(Path.Combine(path, ".git")));
    }

    [Fact]
    public void Init_Bare_CreatesBareRepository()
    {
        var path = Path.Combine(_tmpDir, "bare");
        using var repo = Repository.Init(path, isBare: true);
        var cfg = repo.Config();
        Assert.True(cfg.Core.IsBare);
    }

    // --- Open ---

    [Fact]
    public void Open_AfterInit_Succeeds()
    {
        var path = Path.Combine(_tmpDir, "repo");
        using (Repository.Init(path)) { }
        using var repo = Repository.Open(path);
        Assert.NotNull(repo);
    }

    [Fact]
    public void Open_NonExistentPath_Throws()
    {
        Assert.Throws<GoGitException>(() => Repository.Open(Path.Combine(_tmpDir, "does-not-exist")));
    }

    // --- Config ---

    [Fact]
    public void GetConfig_NonBare_IsBareIsFalse()
    {
        var path = Path.Combine(_tmpDir, "repo");
        using var repo = Repository.Init(path);
        var cfg = repo.Config();
        Assert.False(cfg.Core.IsBare);
    }

    // --- References ---

    [Fact]
    public void SetReference_ThenReference_ReturnsCommitHash()
    {
        var path = Path.Combine(_tmpDir, "bare");
        using var repo = Repository.Init(path, isBare: true);

        var blobHash = repo.StoreBlob("initial\n");
        var treeHash = repo.StoreTree(
        [
            new TreeEntryInfo { Name = "file.txt", Hash = blobHash, Mode = TreeEntryMode.NonExecutableFile }
        ]);
        var commitHash = repo.StoreCommit(
            treeHash, [],
            "Author", "author@test.com",
            "Author", "author@test.com",
            "Initial commit\n",
            new DateTimeOffset(2024, 1, 1, 0, 0, 0, TimeSpan.Zero));

        repo.SetReference("refs/heads/main", commitHash);

        var (_, resolvedHash) = repo.Reference("refs/heads/main", resolved: true);
        Assert.Equal(commitHash, resolvedHash);
    }

    [Fact]
    public void References_AfterSetReference_ContainsRef()
    {
        var path = Path.Combine(_tmpDir, "bare");
        using var repo = Repository.Init(path, isBare: true);

        var blobHash = repo.StoreBlob("content\n");
        var treeHash = repo.StoreTree(
        [
            new TreeEntryInfo { Name = "f.txt", Hash = blobHash, Mode = TreeEntryMode.NonExecutableFile }
        ]);
        var commitHash = repo.StoreCommit(treeHash, [], "A", "a@b.com", "A", "a@b.com", "msg\n", DateTimeOffset.UtcNow);
        repo.SetReference("refs/heads/main", commitHash);

        using var iter = repo.References();
        var names = iter.Select(r => r.Name).ToList();
        Assert.Contains("refs/heads/main", names);
    }
}
