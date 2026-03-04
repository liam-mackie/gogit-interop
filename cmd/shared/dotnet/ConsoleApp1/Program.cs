using GoGit.Interop;

Console.WriteLine("Hello, World!");

using var cloneOpts = new CloneOptions();
var sshKey =
    "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACDLaDxtHPSJlWQJG4P7kYWXpPwDXlIUXlwYYS4phSBUZwAAAKCYKf1SmCn9\nUgAAAAtzc2gtZWQyNTUxOQAAACDLaDxtHPSJlWQJG4P7kYWXpPwDXlIUXlwYYS4phSBUZw\nAAAECly1MyCodTWzyU8U33dE+uVv3KTFzeYwW9lpjlLj2ldMtoPG0c9ImVZAkbg/uRhZek\n/ANeUhReXBhhLimFIFRnAAAAF2xpYW0ubWFja2llQG9jdG9wdXMuY29tAQIDBAUG\n-----END OPENSSH PRIVATE KEY-----\n";
cloneOpts.SetURL("github.com:liam-mackie/stackline-swift.git");
cloneOpts.SetAuth(Auth.SSHKey("git", sshKey));
cloneOpts.SetNoCheckout(true);

using var repo = Repository.Clone("liam-3", cloneOpts);
foreach (var commit in repo.Log())
{
    Console.WriteLine($"{commit.Hash[..8]} {commit.Message.TrimEnd()}");
}
