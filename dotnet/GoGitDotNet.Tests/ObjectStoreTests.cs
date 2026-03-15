using System.Text;
using GoGitDotNet;
using Xunit;

namespace GoGitDotNet.Tests;

/// <summary>
/// Tests for the low-level object store operations: StoreBlob, GetTreeEntries,
/// StoreTree, StoreCommit, and SetReference — covering the full workflow for
/// creating commits without a working tree checkout.
/// </summary>
public class ObjectStoreTests : IDisposable
{
    private readonly string _tmpDir;
    private readonly Repository _repo;

    public ObjectStoreTests()
    {
        _tmpDir = Path.Combine(Path.GetTempPath(), "gogitdotnet-" + Path.GetRandomFileName());
        Directory.CreateDirectory(_tmpDir);
        _repo = Repository.Init(_tmpDir, isBare: true);
    }

    public void Dispose()
    {
        _repo.Dispose();
        TestHelpers.DeleteDirectory(_tmpDir);
    }

    // --- StoreBlob ---

    [Fact]
    public void StoreBlob_String_ReturnsSha1Hash()
    {
        var hash = _repo.StoreBlob("Hello, world!\n");
        Assert.Equal(40, hash.Length);
        Assert.Matches("^[0-9a-f]{40}$", hash);
    }

    [Fact]
    public void StoreBlob_ByteArray_ReturnsSha1Hash()
    {
        var hash = _repo.StoreBlob(Encoding.UTF8.GetBytes("Hello, world!\n"));
        Assert.Equal(40, hash.Length);
    }

    [Fact]
    public void StoreBlob_StringAndBytes_ProduceSameHash()
    {
        var text = "identical content";
        var h1 = _repo.StoreBlob(text);
        var h2 = _repo.StoreBlob(Encoding.UTF8.GetBytes(text));
        Assert.Equal(h1, h2);
    }

    [Fact]
    public void StoreBlob_SameContent_IsDeterministic()
    {
        var h1 = _repo.StoreBlob("deterministic");
        var h2 = _repo.StoreBlob("deterministic");
        Assert.Equal(h1, h2);
    }

    [Fact]
    public void StoreBlob_DifferentContent_ProducesDifferentHash()
    {
        var h1 = _repo.StoreBlob("content A");
        var h2 = _repo.StoreBlob("content B");
        Assert.NotEqual(h1, h2);
    }

    [Fact]
    public void StoreBlob_ThenGetBlobObject_ReturnsOriginalContent()
    {
        var content = "blob round-trip test\n";
        var hash = _repo.StoreBlob(content);
        using var blob = _repo.BlobObject(hash);
        Assert.Equal(content, blob.Contents());
    }

    [Fact]
    public void StoreBlob_EmptyContent_Succeeds()
    {
        var hash = _repo.StoreBlob("");
        Assert.Equal(40, hash.Length);
    }

    // --- StoreTree ---

    [Fact]
    public void StoreTree_SingleEntry_ReturnsSha1Hash()
    {
        var blobHash = _repo.StoreBlob("readme\n");
        var treeHash = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "README.md", Hash = blobHash, Mode = TreeEntryMode.NonExecutableFile }
        ]);
        Assert.Equal(40, treeHash.Length);
        Assert.Matches("^[0-9a-f]{40}$", treeHash);
    }

    [Fact]
    public void StoreTree_SameEntries_IsDeterministic()
    {
        var blobHash = _repo.StoreBlob("content\n");
        var entry = new TreeEntryInfo { Name = "file.txt", Hash = blobHash, Mode = TreeEntryMode.NonExecutableFile };
        var h1 = _repo.StoreTree([entry]);
        var h2 = _repo.StoreTree([entry]);
        Assert.Equal(h1, h2);
    }

    // --- GetTreeEntries ---

    [Fact]
    public void GetTreeEntries_SingleFile_ReturnsEntry()
    {
        var blobHash = _repo.StoreBlob("content\n");
        var treeHash = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "file.txt", Hash = blobHash, Mode = TreeEntryMode.NonExecutableFile }
        ]);

        var entries = _repo.TreeEntries(treeHash);

        Assert.Single(entries);
        Assert.Equal("file.txt", entries[0].Name);
        Assert.Equal(blobHash, entries[0].Hash);
        Assert.Equal(TreeEntryMode.NonExecutableFile, entries[0].Mode);
    }

    [Fact]
    public void GetTreeEntries_MultipleFiles_ReturnsAllEntries()
    {
        var h1 = _repo.StoreBlob("file1\n");
        var h2 = _repo.StoreBlob("file2\n");
        var treeHash = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "a.txt", Hash = h1, Mode = TreeEntryMode.NonExecutableFile },
            new TreeEntryInfo { Name = "b.txt", Hash = h2, Mode = TreeEntryMode.NonExecutableFile },
        ]);

        var entries = _repo.TreeEntries(treeHash);

        Assert.Equal(2, entries.Length);
        Assert.Contains(entries, e => e.Name == "a.txt" && e.Hash == h1);
        Assert.Contains(entries, e => e.Name == "b.txt" && e.Hash == h2);
    }

    [Fact]
    public void GetTreeEntries_ExecutableFile_PreservesMode()
    {
        var blobHash = _repo.StoreBlob("#!/bin/sh\n");
        var treeHash = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "run.sh", Hash = blobHash, Mode = TreeEntryMode.ExecutableFile }
        ]);

        var entries = _repo.TreeEntries(treeHash);
        Assert.Equal(TreeEntryMode.ExecutableFile, entries[0].Mode);
    }

    // --- StoreCommit ---

    [Fact]
    public void StoreCommit_RootCommit_ReturnsSha1Hash()
    {
        var blobHash = _repo.StoreBlob("content\n");
        var treeHash = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "file.txt", Hash = blobHash, Mode = TreeEntryMode.NonExecutableFile }
        ]);
        var when = new DateTimeOffset(2024, 1, 1, 0, 0, 0, TimeSpan.Zero);

        var commitHash = _repo.StoreCommit(
            treeHash, [],
            "Alice", "alice@example.com",
            "Alice", "alice@example.com",
            "Initial commit\n", when);

        Assert.Equal(40, commitHash.Length);
        Assert.Matches("^[0-9a-f]{40}$", commitHash);
    }

    [Fact]
    public void StoreCommit_ThenGetCommitObject_ReturnsCorrectMetadata()
    {
        var blobHash = _repo.StoreBlob("content\n");
        var treeHash = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "file.txt", Hash = blobHash, Mode = TreeEntryMode.NonExecutableFile }
        ]);
        var when = new DateTimeOffset(2024, 6, 15, 12, 0, 0, TimeSpan.Zero);

        var commitHash = _repo.StoreCommit(
            treeHash, [],
            "Bob", "bob@example.com",
            "Bot", "bot@example.com",
            "Add file.txt\n", when);

        using var commit = _repo.CommitObject(commitHash);
        Assert.Equal("Add file.txt\n", commit.Message);
        Assert.Equal("Bob", commit.AuthorName);
        Assert.Equal("bob@example.com", commit.AuthorEmail);
        Assert.Equal("Bot", commit.CommitterName);
        Assert.Equal("bot@example.com", commit.CommitterEmail);
        Assert.Equal(when.ToUnixTimeSeconds(), commit.AuthorWhen.ToUnixTimeSeconds());
        Assert.Equal(0, commit.NumParents());
    }

    [Fact]
    public void StoreCommit_WithParent_RecordsParentHash()
    {
        var blobHash = _repo.StoreBlob("v1\n");
        var treeHash = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "f.txt", Hash = blobHash, Mode = TreeEntryMode.NonExecutableFile }
        ]);
        var when = DateTimeOffset.UtcNow;

        var parent = _repo.StoreCommit(treeHash, [], "A", "a@b.com", "A", "a@b.com", "v1\n", when);

        var blobHash2 = _repo.StoreBlob("v2\n");
        var treeHash2 = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "f.txt", Hash = blobHash2, Mode = TreeEntryMode.NonExecutableFile }
        ]);
        var child = _repo.StoreCommit(treeHash2, [parent], "A", "a@b.com", "A", "a@b.com", "v2\n", when);

        using var commit = _repo.CommitObject(child);
        Assert.Equal(1, commit.NumParents());
        using var parentCommit = commit.Parent(0);
        Assert.Equal(parent, parentCommit.Hash);
    }

    [Fact]
    public void StoreCommit_TreeHashMatchesStoredTree()
    {
        var blobHash = _repo.StoreBlob("data\n");
        var treeHash = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "data.txt", Hash = blobHash, Mode = TreeEntryMode.NonExecutableFile }
        ]);
        var commitHash = _repo.StoreCommit(treeHash, [], "A", "a@b.com", "A", "a@b.com", "msg\n", DateTimeOffset.UtcNow);

        using var commit = _repo.CommitObject(commitHash);
        Assert.Equal(treeHash, commit.TreeHash);
    }

    // --- Full workflow ---

    [Fact]
    public void FullWorkflow_CreateAndReadCommit()
    {
        // 1. Store blob
        var content = "Hello, GoGitDotNet!\n";
        var blobHash = _repo.StoreBlob(content);

        // 2. Build tree
        var treeHash = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "hello.txt", Hash = blobHash, Mode = TreeEntryMode.NonExecutableFile }
        ]);

        // 3. Create commit
        var when = new DateTimeOffset(2024, 6, 1, 12, 0, 0, TimeSpan.Zero);
        var commitHash = _repo.StoreCommit(
            treeHash, [],
            "Alice", "alice@example.com",
            "Alice", "alice@example.com",
            "Add hello.txt\n", when);

        // 4. Advance branch
        _repo.SetReference("refs/heads/main", commitHash);

        // 5. Verify via object store
        using var commit = _repo.CommitObject(commitHash);
        Assert.Equal("Add hello.txt\n", commit.Message);

        using var tree = commit.Tree();
        var entries = _repo.TreeEntries(tree.Hash);
        Assert.Single(entries);
        Assert.Equal("hello.txt", entries[0].Name);

        using var blob = _repo.BlobObject(blobHash);
        Assert.Equal(content, blob.Contents());
    }

    [Fact]
    public void FullWorkflow_ReplaceFileAcrossTwoCommits()
    {
        var when = new DateTimeOffset(2024, 1, 1, 0, 0, 0, TimeSpan.Zero);

        // First commit
        var blob1 = _repo.StoreBlob("version 1\n");
        var tree1 = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "README.md", Hash = blob1, Mode = TreeEntryMode.NonExecutableFile }
        ]);
        var commit1 = _repo.StoreCommit(tree1, [], "A", "a@b.com", "A", "a@b.com", "v1\n", when);

        // Second commit — replace file content
        var blob2 = _repo.StoreBlob("version 2\n");
        var entries = _repo.TreeEntries(tree1).ToList();
        entries[0] = new TreeEntryInfo { Name = entries[0].Name, Hash = blob2, Mode = entries[0].Mode };
        var tree2 = _repo.StoreTree(entries);
        var commit2 = _repo.StoreCommit(tree2, [commit1], "A", "a@b.com", "A", "a@b.com", "v2\n", when);

        _repo.SetReference("refs/heads/main", commit2);

        // Verify second commit's blob
        using var c2 = _repo.CommitObject(commit2);
        using var t2 = c2.Tree();
        var entry = _repo.TreeEntries(t2.Hash)[0];
        using var blob = _repo.BlobObject(entry.Hash);
        Assert.Equal("version 2\n", blob.Contents());

        // Verify parent chain
        Assert.Equal(1, c2.NumParents());
        using var p = c2.Parent(0);
        Assert.Equal(commit1, p.Hash);
    }

    // --- Add / Edit / Remove without checkout ---

    [Fact]
    public void NoCheckout_AddFile_AppearsInNextCommit()
    {
        var when = DateTimeOffset.UtcNow;

        // Seed: one file
        var seedBlob = _repo.StoreBlob("original\n");
        var seedTree = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "existing.txt", Hash = seedBlob, Mode = TreeEntryMode.NonExecutableFile }
        ]);
        var seedCommit = _repo.StoreCommit(seedTree, [], "A", "a@b.com", "A", "a@b.com", "seed\n", when);
        _repo.SetReference("refs/heads/main", seedCommit);

        // Add a new file by mutating the entry list
        var newBlob = _repo.StoreBlob("brand new\n");
        var entries = _repo.TreeEntries(seedTree).ToList();
        entries.Add(new TreeEntryInfo { Name = "added.txt", Hash = newBlob, Mode = TreeEntryMode.NonExecutableFile });
        var newTree = _repo.StoreTree(entries);
        var newCommit = _repo.StoreCommit(newTree, [seedCommit], "A", "a@b.com", "A", "a@b.com", "add file\n", when);
        _repo.SetReference("refs/heads/main", newCommit);

        using var c = _repo.CommitObject(newCommit);
        using var t = c.Tree();
        var result = _repo.TreeEntries(t.Hash);
        Assert.Equal(2, result.Length);
        var added = result.Single(e => e.Name == "added.txt");
        using var blob = _repo.BlobObject(added.Hash);
        Assert.Equal("brand new\n", blob.Contents());
    }

    [Fact]
    public void NoCheckout_EditFile_ContentUpdatedInNextCommit()
    {
        var when = DateTimeOffset.UtcNow;

        var v1Blob = _repo.StoreBlob("version 1\n");
        var tree1 = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "file.txt", Hash = v1Blob, Mode = TreeEntryMode.NonExecutableFile }
        ]);
        var commit1 = _repo.StoreCommit(tree1, [], "A", "a@b.com", "A", "a@b.com", "v1\n", when);

        // Edit: replace blob for the existing entry
        var v2Blob = _repo.StoreBlob("version 2\n");
        var entries = _repo.TreeEntries(tree1).ToList();
        var i = entries.FindIndex(e => e.Name == "file.txt");
        entries[i] = new TreeEntryInfo { Name = entries[i].Name, Hash = v2Blob, Mode = entries[i].Mode };
        var tree2 = _repo.StoreTree(entries);
        var commit2 = _repo.StoreCommit(tree2, [commit1], "A", "a@b.com", "A", "a@b.com", "v2\n", when);

        using var c = _repo.CommitObject(commit2);
        using var t = c.Tree();
        var entry = _repo.TreeEntries(t.Hash).Single(e => e.Name == "file.txt");
        using var blob = _repo.BlobObject(entry.Hash);
        Assert.Equal("version 2\n", blob.Contents());
    }

    [Fact]
    public void NoCheckout_RemoveFile_AbsentFromNextCommit()
    {
        var when = DateTimeOffset.UtcNow;

        var b1 = _repo.StoreBlob("keep\n");
        var b2 = _repo.StoreBlob("delete me\n");
        var tree1 = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "keep.txt",   Hash = b1, Mode = TreeEntryMode.NonExecutableFile },
            new TreeEntryInfo { Name = "delete.txt", Hash = b2, Mode = TreeEntryMode.NonExecutableFile },
        ]);
        var commit1 = _repo.StoreCommit(tree1, [], "A", "a@b.com", "A", "a@b.com", "seed\n", when);

        // Remove: drop the unwanted entry
        var entries = _repo.TreeEntries(tree1).Where(e => e.Name != "delete.txt").ToList();
        var tree2 = _repo.StoreTree(entries);
        var commit2 = _repo.StoreCommit(tree2, [commit1], "A", "a@b.com", "A", "a@b.com", "remove\n", when);

        using var c = _repo.CommitObject(commit2);
        using var t = c.Tree();
        var result = _repo.TreeEntries(t.Hash);
        Assert.Single(result);
        Assert.Equal("keep.txt", result[0].Name);
    }

    [Fact]
    public void NoCheckout_AddEditRemove_AllInOneCommit()
    {
        var when = DateTimeOffset.UtcNow;

        // Seed: two files
        var keepBlob    = _repo.StoreBlob("keep\n");
        var editBlob    = _repo.StoreBlob("original\n");
        var removeBlob  = _repo.StoreBlob("bye\n");
        var seedTree = _repo.StoreTree(
        [
            new TreeEntryInfo { Name = "keep.txt",   Hash = keepBlob,   Mode = TreeEntryMode.NonExecutableFile },
            new TreeEntryInfo { Name = "edit.txt",   Hash = editBlob,   Mode = TreeEntryMode.NonExecutableFile },
            new TreeEntryInfo { Name = "remove.txt", Hash = removeBlob, Mode = TreeEntryMode.NonExecutableFile },
        ]);
        var seedCommit = _repo.StoreCommit(seedTree, [], "A", "a@b.com", "A", "a@b.com", "seed\n", when);

        // Mutate: add + edit + remove in a single tree
        var newBlob     = _repo.StoreBlob("new file\n");
        var updatedBlob = _repo.StoreBlob("updated\n");

        var entries = _repo.TreeEntries(seedTree).ToList();
        entries.RemoveAll(e => e.Name == "remove.txt");
        var ei = entries.FindIndex(e => e.Name == "edit.txt");
        entries[ei] = new TreeEntryInfo { Name = entries[ei].Name, Hash = updatedBlob, Mode = entries[ei].Mode };
        entries.Add(new TreeEntryInfo { Name = "new.txt", Hash = newBlob, Mode = TreeEntryMode.NonExecutableFile });

        var newTree   = _repo.StoreTree(entries);
        var newCommit = _repo.StoreCommit(newTree, [seedCommit], "A", "a@b.com", "A", "a@b.com", "mutate\n", when);
        _repo.SetReference("refs/heads/main", newCommit);

        using var c = _repo.CommitObject(newCommit);
        using var t = c.Tree();
        var result = _repo.TreeEntries(t.Hash);

        Assert.Equal(3, result.Length);
        Assert.Contains(result, e => e.Name == "keep.txt");
        Assert.DoesNotContain(result, e => e.Name == "remove.txt");

        var editEntry = result.Single(e => e.Name == "edit.txt");
        using var editResult = _repo.BlobObject(editEntry.Hash);
        Assert.Equal("updated\n", editResult.Contents());

        var newEntry = result.Single(e => e.Name == "new.txt");
        using var newResult = _repo.BlobObject(newEntry.Hash);
        Assert.Equal("new file\n", newResult.Contents());
    }
}
