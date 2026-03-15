namespace GoGitDotNet.Tests;

internal static class TestHelpers
{
    /// <summary>
    /// Deletes a directory tree, clearing read-only attributes first.
    /// Required on Windows where git object files are written as read-only.
    /// </summary>
    internal static void DeleteDirectory(string path)
    {
        foreach (var file in Directory.GetFiles(path, "*", SearchOption.AllDirectories))
            System.IO.File.SetAttributes(file, FileAttributes.Normal);
        Directory.Delete(path, recursive: true);
    }
}
