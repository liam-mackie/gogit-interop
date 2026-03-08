namespace GoGitDotNet;

/// <summary>Controls which tags are fetched during a clone or fetch operation.</summary>
public enum TagFetchMode
{
    Invalid = 0,
    All = 1,
    No = 2,
    Follow = 3,
    FollowAny = 4,
}
