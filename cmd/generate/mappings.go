package main

import "fmt"

var handleTypes = map[string]string{
	"*git.Repository": "Repository",
	"*git.Worktree":   "Worktree",
	"*git.Remote":     "Remote",
	"*git.Submodule":  "Submodule",
}

var optionsTypes = map[string]string{
	"*git.CloneOptions":           "CloneOptions",
	"*git.PullOptions":            "PullOptions",
	"*git.FetchOptions":           "FetchOptions",
	"*git.PushOptions":            "PushOptions",
	"*git.CheckoutOptions":        "CheckoutOptions",
	"*git.ResetOptions":           "ResetOptions",
	"*git.RestoreOptions":         "RestoreOptions",
	"*git.CommitOptions":          "CommitOptions",
	"*git.CreateTagOptions":       "CreateTagOptions",
	"*git.AddOptions":             "AddOptions",
	"*git.CleanOptions":           "CleanOptions",
	"*git.GrepOptions":            "GrepOptions",
	"*git.LogOptions":             "LogOptions",
	"*git.PlainOpenOptions":       "PlainOpenOptions",
	"*git.ListOptions":            "ListOptions",
	"*git.MergeOptions":           "MergeOptions",
	"*git.SubmoduleUpdateOptions": "SubmoduleUpdateOptions",
	"git.StatusOptions":           "StatusOptions",
}

var iteratorTypes = map[string]iteratorInfo{
	"object.CommitIter":    {ItemType: "Commit", NextMethod: "Next", CloseMethod: "Close"},
	"storer.ReferenceIter": {ItemType: "Reference", NextMethod: "Next", CloseMethod: "Close"},
}

type iteratorInfo struct {
	ItemType    string
	NextMethod  string
	CloseMethod string
}

func resolveMapping(goType string) TypeMapping {
	if name, ok := handleTypes[goType]; ok {
		return TypeMapping{
			Kind:       MappingHandle,
			GoType:     goType,
			CType:      "C.longlong",
			CSharpType: "long",
			HandleType: name,
		}
	}

	if name, ok := optionsTypes[goType]; ok {
		return TypeMapping{
			Kind:       MappingOptions,
			GoType:     goType,
			CType:      "C.longlong",
			CSharpType: "long",
			HandleType: name,
		}
	}

	switch goType {
	case "string":
		return TypeMapping{Kind: MappingString, GoType: "string", CType: "*C.char", CSharpType: "string"}
	case "bool":
		return TypeMapping{Kind: MappingBool, GoType: "bool", CType: "C.int", CSharpType: "int"}
	case "int", "int32":
		return TypeMapping{Kind: MappingPrimitive, GoType: goType, CType: "C.int", CSharpType: "int"}
	case "int64":
		return TypeMapping{Kind: MappingPrimitive, GoType: "int64", CType: "C.longlong", CSharpType: "long"}
	case "uint", "uint32":
		return TypeMapping{Kind: MappingPrimitive, GoType: goType, CType: "C.uint", CSharpType: "uint"}
	case "int8":
		return TypeMapping{Kind: MappingPrimitive, GoType: "int8", CType: "C.int", CSharpType: "int"}
	case "uint8":
		return TypeMapping{Kind: MappingPrimitive, GoType: "uint8", CType: "C.int", CSharpType: "int"}
	case "[]byte":
		return TypeMapping{Kind: MappingByteSlice, GoType: "[]byte", CType: "*C.char", CSharpType: "byte[]"}
	case "error":
		return TypeMapping{Kind: MappingPrimitive, GoType: "error", CType: "*C.char", CSharpType: "IntPtr"}
	case "plumbing.Hash":
		return TypeMapping{Kind: MappingHash, GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string"}
	case "plumbing.ReferenceName":
		return TypeMapping{Kind: MappingReferenceName, GoType: "plumbing.ReferenceName", CType: "*C.char", CSharpType: "string"}
	case "transport.AuthMethod":
		return TypeMapping{Kind: MappingAuth, GoType: "transport.AuthMethod", CType: "C.longlong", CSharpType: "long"}
	case "git.Signer":
		return TypeMapping{Kind: MappingSigner, GoType: "git.Signer", CType: "C.longlong", CSharpType: "long"}
	case "sideband.Progress":
		return TypeMapping{Kind: MappingCallback, GoType: "sideband.Progress", CType: "C.longlong", CSharpType: "long"}
	case "object.CommitIter":
		return TypeMapping{Kind: MappingIterator, GoType: "object.CommitIter", CType: "C.longlong", CSharpType: "long", HandleType: "CommitIter"}
	case "storer.ReferenceIter":
		return TypeMapping{Kind: MappingIterator, GoType: "storer.ReferenceIter", CType: "C.longlong", CSharpType: "long", HandleType: "ReferenceIter"}
	case "plumbing.TagMode":
		return TypeMapping{Kind: MappingEnum, GoType: "plumbing.TagMode", CType: "C.int", CSharpType: "int"}
	case "git.ResetMode":
		return TypeMapping{Kind: MappingEnum, GoType: "git.ResetMode", CType: "C.int", CSharpType: "int"}
	case "git.LogOrder":
		return TypeMapping{Kind: MappingEnum, GoType: "git.LogOrder", CType: "C.int", CSharpType: "int"}
	case "git.MergeStrategy":
		return TypeMapping{Kind: MappingEnum, GoType: "git.MergeStrategy", CType: "C.int", CSharpType: "int"}
	case "git.SubmoduleRecursivity":
		return TypeMapping{Kind: MappingEnum, GoType: "git.SubmoduleRecursivity", CType: "C.uint", CSharpType: "uint"}
	case "git.PeelingOption":
		return TypeMapping{Kind: MappingEnum, GoType: "git.PeelingOption", CType: "C.int", CSharpType: "int"}
	case "git.StatusStrategy":
		return TypeMapping{Kind: MappingEnum, GoType: "git.StatusStrategy", CType: "C.int", CSharpType: "int"}
	}

	return TypeMapping{Kind: MappingSkip, GoType: goType}
}

func isMappable(goType string) bool {
	return resolveMapping(goType).Kind != MappingSkip
}

func warnUnmappable(context, goType string) {
	fmt.Printf("WARNING: skipping %s — unmappable type %s\n", context, goType)
}
