package main

func analyze() *Package {
	pkg := &Package{Name: "gogit"}

	pkg.Functions = buildTopLevelFunctions()
	pkg.Types = buildHandleTypes()
	pkg.Options = buildOptionsStructs()

	return pkg
}

func buildTopLevelFunctions() []Function {
	return []Function{
		{
			GoName: "PlainInit",
			CName:  "GitPlainInit",
			Params: []Param{
				{GoName: "path", GoType: "string", CName: "path", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
				{GoName: "isBare", GoType: "bool", CName: "isBare", CType: "C.int", CSharpType: "int", Mapping: resolveMapping("bool")},
			},
			Returns: []Return{
				{GoType: "*git.Repository", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("*git.Repository")},
				{GoType: "error", CType: "*C.char", CSharpType: "IntPtr", Mapping: resolveMapping("error"), IsError: true},
			},
		},
		{
			GoName: "PlainOpen",
			CName:  "GitPlainOpen",
			Params: []Param{
				{GoName: "path", GoType: "string", CName: "path", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
			},
			Returns: []Return{
				{GoType: "*git.Repository", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("*git.Repository")},
				{GoType: "error", CType: "*C.char", CSharpType: "IntPtr", Mapping: resolveMapping("error"), IsError: true},
			},
		},
		{
			GoName: "PlainOpenWithOptions",
			CName:  "GitPlainOpenWithOptions",
			Params: []Param{
				{GoName: "path", GoType: "string", CName: "path", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
				{GoName: "opts", GoType: "*git.PlainOpenOptions", CName: "optsHandle", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("*git.PlainOpenOptions")},
			},
			Returns: []Return{
				{GoType: "*git.Repository", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("*git.Repository")},
				{GoType: "error", CType: "*C.char", CSharpType: "IntPtr", Mapping: resolveMapping("error"), IsError: true},
			},
		},
		{
			GoName: "PlainCloneWithOptions",
			CName:  "GitPlainClone",
			Params: []Param{
				{GoName: "path", GoType: "string", CName: "path", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
				{GoName: "opts", GoType: "*git.CloneOptions", CName: "optsHandle", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("*git.CloneOptions")},
			},
			Returns: []Return{
				{GoType: "*git.Repository", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("*git.Repository")},
				{GoType: "error", CType: "*C.char", CSharpType: "IntPtr", Mapping: resolveMapping("error"), IsError: true},
			},
		},
	}
}

func buildHandleTypes() []HandleType {
	return []HandleType{
		buildRepositoryType(),
		buildWorktreeType(),
		buildRemoteType(),
		buildSubmoduleType(),
	}
}

func buildRepositoryType() HandleType {
	return HandleType{
		GoName:    "Repository",
		CPrefix:   "GitRepository",
		IsPointer: true,
		Methods: []Method{
			simpleMethod("Head", "GitRepositoryHead", "Repository", nil, []Return{
				{GoType: "plumbing.ReferenceName", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.ReferenceName")},
				{GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Worktree", "GitRepositoryWorktree", "Repository", nil, []Return{
				{GoType: "*git.Worktree", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("*git.Worktree")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Fetch", "GitRepositoryFetch", "Repository", []Param{
				optionsParam("opts", "*git.FetchOptions", "FetchOptions"),
			}, errorReturn()),
			simpleMethod("Push", "GitRepositoryPush", "Repository", []Param{
				optionsParam("opts", "*git.PushOptions", "PushOptions"),
			}, errorReturn()),
			simpleMethod("Log", "GitRepositoryLog", "Repository", []Param{
				optionsParam("opts", "*git.LogOptions", "LogOptions"),
			}, []Return{
				{GoType: "object.CommitIter", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("object.CommitIter")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Tags", "GitRepositoryTags", "Repository", nil, []Return{
				{GoType: "storer.ReferenceIter", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("storer.ReferenceIter")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Branches", "GitRepositoryBranches", "Repository", nil, []Return{
				{GoType: "storer.ReferenceIter", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("storer.ReferenceIter")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Notes", "GitRepositoryNotes", "Repository", nil, []Return{
				{GoType: "storer.ReferenceIter", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("storer.ReferenceIter")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("References", "GitRepositoryReferences", "Repository", nil, []Return{
				{GoType: "storer.ReferenceIter", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("storer.ReferenceIter")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Reference", "GitRepositoryReference", "Repository", []Param{
				stringParam("name", "plumbing.ReferenceName"),
				boolParam("resolved"),
			}, []Return{
				{GoType: "plumbing.ReferenceName", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.ReferenceName")},
				{GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("ResolveRevision", "GitRepositoryResolveRevision", "Repository", []Param{
				{GoName: "rev", GoType: "string", CName: "rev", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
			}, []Return{
				{GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("CreateRemote", "GitRepositoryCreateRemote", "Repository", []Param{
				stringParam("name", "string"),
				{GoName: "url", GoType: "string", CName: "url", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
			}, []Return{
				{GoType: "*git.Remote", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("*git.Remote")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Remote", "GitRepositoryRemote", "Repository", []Param{
				stringParam("name", "string"),
			}, []Return{
				{GoType: "*git.Remote", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("*git.Remote")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("DeleteRemote", "GitRepositoryDeleteRemote", "Repository", []Param{
				stringParam("name", "string"),
			}, errorReturn()),
			simpleMethod("CreateBranch", "GitRepositoryCreateBranch", "Repository", []Param{
				stringParam("name", "string"),
				{GoName: "hash", GoType: "plumbing.Hash", CName: "hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
			}, errorReturn()),
			simpleMethod("DeleteBranch", "GitRepositoryDeleteBranch", "Repository", []Param{
				stringParam("name", "string"),
			}, errorReturn()),
			simpleMethod("CreateTag", "GitRepositoryCreateTag", "Repository", []Param{
				stringParam("name", "string"),
				{GoName: "hash", GoType: "plumbing.Hash", CName: "hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
				optionsParam("opts", "*git.CreateTagOptions", "CreateTagOptions"),
			}, []Return{
				{GoType: "plumbing.ReferenceName", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.ReferenceName")},
				{GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Tag", "GitRepositoryTag", "Repository", []Param{
				stringParam("name", "string"),
			}, []Return{
				{GoType: "plumbing.ReferenceName", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.ReferenceName")},
				{GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("DeleteTag", "GitRepositoryDeleteTag", "Repository", []Param{
				stringParam("name", "string"),
			}, errorReturn()),
			simpleMethod("CommitObject", "GitRepositoryCommitObject", "Repository", []Param{
				{GoName: "hash", GoType: "plumbing.Hash", CName: "hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
			}, []Return{
				{GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
				{GoType: "string", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
				{GoType: "string", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
				{GoType: "string", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
				{GoType: "int64", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("int64")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Merge", "GitRepositoryMerge", "Repository", []Param{
				stringParam("refName", "plumbing.ReferenceName"),
				{GoName: "hash", GoType: "plumbing.Hash", CName: "hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
				optionsParam("opts", "*git.MergeOptions", "MergeOptions"),
			}, errorReturn()),
		},
	}
}

func buildWorktreeType() HandleType {
	return HandleType{
		GoName:    "Worktree",
		CPrefix:   "GitWorktree",
		IsPointer: true,
		Methods: []Method{
			simpleMethod("Status", "GitWorktreeStatus", "Worktree", nil, []Return{
				{GoType: "string", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Add", "GitWorktreeAdd", "Worktree", []Param{
				{GoName: "path", GoType: "string", CName: "path", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
			}, []Return{
				{GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("AddWithOptions", "GitWorktreeAddWithOptions", "Worktree", []Param{
				optionsParam("opts", "*git.AddOptions", "AddOptions"),
			}, errorReturn()),
			simpleMethod("AddGlob", "GitWorktreeAddGlob", "Worktree", []Param{
				{GoName: "pattern", GoType: "string", CName: "pattern", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
			}, errorReturn()),
			simpleMethod("Commit", "GitWorktreeCommit", "Worktree", []Param{
				{GoName: "msg", GoType: "string", CName: "msg", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
				optionsParam("opts", "*git.CommitOptions", "CommitOptions"),
			}, []Return{
				{GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Checkout", "GitWorktreeCheckout", "Worktree", []Param{
				optionsParam("opts", "*git.CheckoutOptions", "CheckoutOptions"),
			}, errorReturn()),
			simpleMethod("Pull", "GitWorktreePull", "Worktree", []Param{
				optionsParam("opts", "*git.PullOptions", "PullOptions"),
			}, errorReturn()),
			simpleMethod("Reset", "GitWorktreeReset", "Worktree", []Param{
				optionsParam("opts", "*git.ResetOptions", "ResetOptions"),
			}, errorReturn()),
			simpleMethod("Restore", "GitWorktreeRestore", "Worktree", []Param{
				optionsParam("opts", "*git.RestoreOptions", "RestoreOptions"),
			}, errorReturn()),
			simpleMethod("Clean", "GitWorktreeClean", "Worktree", []Param{
				optionsParam("opts", "*git.CleanOptions", "CleanOptions"),
			}, errorReturn()),
			simpleMethod("Move", "GitWorktreeMove", "Worktree", []Param{
				{GoName: "from", GoType: "string", CName: "fromPath", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
				{GoName: "to", GoType: "string", CName: "toPath", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
			}, []Return{
				{GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("plumbing.Hash")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Submodule", "GitWorktreeSubmodule", "Worktree", []Param{
				stringParam("name", "string"),
			}, []Return{
				{GoType: "*git.Submodule", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("*git.Submodule")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
			simpleMethod("Submodules", "GitWorktreeSubmodules", "Worktree", nil, []Return{
				{GoType: "string", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
		},
	}
}

func buildRemoteType() HandleType {
	return HandleType{
		GoName:    "Remote",
		CPrefix:   "GitRemote",
		IsPointer: true,
		Methods: []Method{
			simpleMethod("Fetch", "GitRemoteFetch", "Remote", []Param{
				optionsParam("opts", "*git.FetchOptions", "FetchOptions"),
			}, errorReturn()),
			simpleMethod("Push", "GitRemotePush", "Remote", []Param{
				optionsParam("opts", "*git.PushOptions", "PushOptions"),
			}, errorReturn()),
			simpleMethod("List", "GitRemoteList", "Remote", []Param{
				optionsParam("opts", "*git.ListOptions", "ListOptions"),
			}, []Return{
				{GoType: "string", CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
		},
	}
}

func buildSubmoduleType() HandleType {
	return HandleType{
		GoName:    "Submodule",
		CPrefix:   "GitSubmodule",
		IsPointer: true,
		Methods: []Method{
			simpleMethod("Init", "GitSubmoduleInit", "Submodule", nil, errorReturn()),
			simpleMethod("Update", "GitSubmoduleUpdate", "Submodule", []Param{
				optionsParam("opts", "*git.SubmoduleUpdateOptions", "SubmoduleUpdateOptions"),
			}, errorReturn()),
			simpleMethod("Repository", "GitSubmoduleRepository", "Submodule", nil, []Return{
				{GoType: "*git.Repository", CType: "C.longlong", CSharpType: "long", Mapping: resolveMapping("*git.Repository")},
				{GoType: "error", IsError: true, Mapping: resolveMapping("error")},
			}),
		},
	}
}

func buildOptionsStructs() []OptionsStruct {
	return []OptionsStruct{
		{
			GoName: "CloneOptions", CPrefix: "GitCloneOptions",
			Fields: []OptionsField{
				stringField("URL", "url"),
				authField("Auth"),
				stringField("RemoteName", "remoteName"),
				refNameField("ReferenceName", "referenceName"),
				boolField("SingleBranch", "singleBranch"),
				boolField("Mirror", "mirror"),
				boolField("NoCheckout", "noCheckout"),
				intField("Depth", "depth"),
				boolField("Bare", "bare"),
				boolField("InsecureSkipTLS", "insecureSkipTLS"),
				boolField("Shared", "shared"),
				enumField("Tags", "tags", "plumbing.TagMode"),
			},
		},
		{
			GoName: "PullOptions", CPrefix: "GitPullOptions",
			Fields: []OptionsField{
				stringField("RemoteName", "remoteName"),
				stringField("RemoteURL", "remoteURL"),
				refNameField("ReferenceName", "referenceName"),
				boolField("SingleBranch", "singleBranch"),
				intField("Depth", "depth"),
				authField("Auth"),
				boolField("Force", "force"),
				boolField("InsecureSkipTLS", "insecureSkipTLS"),
			},
		},
		{
			GoName: "FetchOptions", CPrefix: "GitFetchOptions",
			Fields: []OptionsField{
				stringField("RemoteName", "remoteName"),
				stringField("RemoteURL", "remoteURL"),
				intField("Depth", "depth"),
				authField("Auth"),
				enumField("Tags", "tags", "plumbing.TagMode"),
				boolField("Force", "force"),
				boolField("InsecureSkipTLS", "insecureSkipTLS"),
				boolField("Prune", "prune"),
			},
		},
		{
			GoName: "PushOptions", CPrefix: "GitPushOptions",
			Fields: []OptionsField{
				stringField("RemoteName", "remoteName"),
				stringField("RemoteURL", "remoteURL"),
				authField("Auth"),
				boolField("Prune", "prune"),
				boolField("Force", "force"),
				boolField("InsecureSkipTLS", "insecureSkipTLS"),
				boolField("FollowTags", "followTags"),
				boolField("Atomic", "atomic"),
				boolField("Quiet", "quiet"),
			},
		},
		{
			GoName: "CheckoutOptions", CPrefix: "GitCheckoutOptions",
			Fields: []OptionsField{
				hashField("Hash", "hash"),
				refNameField("Branch", "branch"),
				boolField("Create", "create"),
				boolField("Force", "force"),
				boolField("Keep", "keep"),
			},
		},
		{
			GoName: "ResetOptions", CPrefix: "GitResetOptions",
			Fields: []OptionsField{
				hashField("Commit", "commit"),
				enumField("Mode", "mode", "git.ResetMode"),
			},
		},
		{
			GoName: "RestoreOptions", CPrefix: "GitRestoreOptions",
			Fields: []OptionsField{
				boolField("Staged", "staged"),
				boolField("Worktree", "worktree"),
			},
		},
		{
			GoName: "CommitOptions", CPrefix: "GitCommitOptions",
			Fields: []OptionsField{
				boolField("All", "all"),
				boolField("AllowEmptyCommits", "allowEmptyCommits"),
				signerField("Signer"),
				boolField("Amend", "amend"),
			},
		},
		{
			GoName: "CreateTagOptions", CPrefix: "GitCreateTagOptions",
			Fields: []OptionsField{
				stringField("Message", "message"),
			},
		},
		{
			GoName: "AddOptions", CPrefix: "GitAddOptions",
			Fields: []OptionsField{
				boolField("All", "all"),
				stringField("Path", "path"),
				stringField("Glob", "glob"),
				boolField("SkipStatus", "skipStatus"),
			},
		},
		{
			GoName: "CleanOptions", CPrefix: "GitCleanOptions",
			Fields: []OptionsField{
				boolField("Dir", "dir"),
			},
		},
		{
			GoName: "LogOptions", CPrefix: "GitLogOptions",
			Fields: []OptionsField{
				hashField("From", "from"),
				hashField("To", "to"),
				enumField("Order", "order", "git.LogOrder"),
				boolField("All", "all"),
			},
		},
		{
			GoName: "PlainOpenOptions", CPrefix: "GitPlainOpenOptions",
			Fields: []OptionsField{
				boolField("DetectDotGit", "detectDotGit"),
			},
		},
		{
			GoName: "ListOptions", CPrefix: "GitListOptions",
			Fields: []OptionsField{
				authField("Auth"),
				boolField("InsecureSkipTLS", "insecureSkipTLS"),
				intField("Timeout", "timeout"),
			},
		},
		{
			GoName: "MergeOptions", CPrefix: "GitMergeOptions",
			Fields: []OptionsField{
				enumField("Strategy", "strategy", "git.MergeStrategy"),
			},
		},
		{
			GoName: "SubmoduleUpdateOptions", CPrefix: "GitSubmoduleUpdateOptions",
			Fields: []OptionsField{
				boolField("Init", "init"),
				boolField("NoFetch", "noFetch"),
				authField("Auth"),
				intField("Depth", "depth"),
			},
		},
	}
}

func simpleMethod(goName, cName, receiver string, params []Param, returns []Return) Method {
	return Method{GoName: goName, CName: cName, Receiver: receiver, Params: params, Returns: returns}
}

func errorReturn() []Return {
	return []Return{{GoType: "error", IsError: true, Mapping: resolveMapping("error")}}
}

func stringParam(name, goType string) Param {
	return Param{GoName: name, GoType: goType, CName: name, CType: "*C.char", CSharpType: "string", Mapping: resolveMapping("string")}
}

func boolParam(name string) Param {
	return Param{GoName: name, GoType: "bool", CName: name, CType: "C.int", CSharpType: "int", Mapping: resolveMapping("bool")}
}

func optionsParam(name, goType, handleName string) Param {
	m := resolveMapping(goType)
	return Param{GoName: name, GoType: goType, CName: name + "Handle", CType: "C.longlong", CSharpType: "long", Mapping: m}
}

func stringField(goName, cSuffix string) OptionsField {
	m := resolveMapping("string")
	return OptionsField{GoName: goName, GoType: "string", CSetterName: cSuffix, CType: "*C.char", CSharpType: "string", Mapping: m}
}

func boolField(goName, cSuffix string) OptionsField {
	m := resolveMapping("bool")
	return OptionsField{GoName: goName, GoType: "bool", CSetterName: cSuffix, CType: "C.int", CSharpType: "bool", Mapping: m}
}

func intField(goName, cSuffix string) OptionsField {
	m := resolveMapping("int")
	return OptionsField{GoName: goName, GoType: "int", CSetterName: cSuffix, CType: "C.int", CSharpType: "int", Mapping: m}
}

func hashField(goName, cSuffix string) OptionsField {
	m := resolveMapping("plumbing.Hash")
	return OptionsField{GoName: goName, GoType: "plumbing.Hash", CSetterName: cSuffix, CType: "*C.char", CSharpType: "string", Mapping: m}
}

func refNameField(goName, cSuffix string) OptionsField {
	m := resolveMapping("plumbing.ReferenceName")
	return OptionsField{GoName: goName, GoType: "plumbing.ReferenceName", CSetterName: cSuffix, CType: "*C.char", CSharpType: "string", Mapping: m}
}

func authField(goName string) OptionsField {
	m := resolveMapping("transport.AuthMethod")
	return OptionsField{GoName: goName, GoType: "transport.AuthMethod", CSetterName: "auth", CType: "C.longlong", CSharpType: "long", Mapping: m}
}

func signerField(goName string) OptionsField {
	m := resolveMapping("git.Signer")
	return OptionsField{GoName: goName, GoType: "git.Signer", CSetterName: "signer", CType: "C.longlong", CSharpType: "long", Mapping: m}
}

func enumField(goName, cSuffix, goType string) OptionsField {
	m := resolveMapping(goType)
	return OptionsField{GoName: goName, GoType: goType, CSetterName: cSuffix, CType: "C.int", CSharpType: "int", Mapping: m}
}
