using Xunit;

namespace GoGitDotNet.Tests;

public class AuthTests
{
    [Fact]
    public void BasicHTTP_Constructs()
    {
        using var auth = Auth.BasicHTTP("user", "password");
        Assert.NotNull(auth);
    }

    [Fact]
    public void TokenHTTP_Constructs()
    {
        using var auth = Auth.TokenHTTP("my-token");
        Assert.NotNull(auth);
    }

    [Fact]
    public void SSHPassword_Constructs()
    {
        using var auth = Auth.SSHPassword("git", "password");
        Assert.NotNull(auth);
    }

    [Fact]
    public void BasicHTTP_DisposeIsIdempotent()
    {
        var auth = Auth.BasicHTTP("user", "pass");
        auth.Dispose();
        auth.Dispose(); // should not throw
    }
}
