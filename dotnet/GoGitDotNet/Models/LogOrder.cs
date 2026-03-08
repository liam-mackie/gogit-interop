namespace GoGitDotNet;

/// <summary>The traversal order for <c>git log</c> iteration.</summary>
public enum LogOrder
{
    Default = 0,
    DFS = 1,
    DFSPost = 2,
    BFS = 3,
    CommitterTime = 4,
}
