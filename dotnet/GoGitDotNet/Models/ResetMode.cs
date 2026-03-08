namespace GoGitDotNet;

/// <summary>The reset mode for a <c>git reset</c> operation.</summary>
public enum ResetMode
{
    Soft = 1,
    Mixed = 2,
    Hard = 3,
    Merge = 4,
    Keep = 5,
}
