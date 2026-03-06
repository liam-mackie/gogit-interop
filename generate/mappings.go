package main

import (
	"fmt"
	"go/types"
	"sync"
)

var (
	dynamicHandles = map[string]string{}
	dynamicOptions = map[string]string{}
	dynamicEnums   = map[string]bool{}
	mappingMu      sync.Mutex
)

var knownPointerIterators = map[string]string{
	"github.com/go-git/go-git/v6/plumbing/object.FileIter": "FileIter",
	"github.com/go-git/go-git/v6/plumbing/object.TreeIter": "TreeIter",
	"github.com/go-git/go-git/v6/plumbing/object.BlobIter": "BlobIter",
	"github.com/go-git/go-git/v6/plumbing/object.TagIter":  "TagIter",
}

func registerHandleMapping(qname, handleName string) {
	mappingMu.Lock()
	defer mappingMu.Unlock()
	dynamicHandles[qname] = handleName
}

func registerOptionsMapping(qname, optionsName string) {
	mappingMu.Lock()
	defer mappingMu.Unlock()
	dynamicOptions[qname] = optionsName
}

func registerEnumMapping(qname string) {
	mappingMu.Lock()
	defer mappingMu.Unlock()
	dynamicEnums[qname] = true
}

func resolveTypeMapping(t types.Type) TypeMapping {
	if alias, ok := t.(*types.Alias); ok {
		return resolveTypeMapping(types.Unalias(alias))
	}

	switch v := t.(type) {
	case *types.Basic:
		return resolveBasicMapping(v)

	case *types.Pointer:
		elem := v.Elem()
		if basic, ok := elem.(*types.Basic); ok && basic.Kind() == types.String {
			return TypeMapping{Kind: MappingString, GoType: "*string", CType: "*C.char", CSharpType: "string"}
		}
		elemResolved := elem
		if alias, ok := elem.(*types.Alias); ok {
			elemResolved = types.Unalias(alias)
		}
		named, ok := elemResolved.(*types.Named)
		if !ok {
			return TypeMapping{Kind: MappingSkip, GoType: t.String()}
		}
		pkg := named.Obj().Pkg()
		if pkg == nil {
			return TypeMapping{Kind: MappingSkip, GoType: t.String()}
		}

		ptrQName := pkg.Path() + "." + named.Obj().Name()
		if ptrQName == "time.Time" {
			return TypeMapping{Kind: MappingTime, GoType: "*time.Time", CType: "C.longlong", CSharpType: "long", QualifiedName: "time.*Time"}
		}
		if ptrQName == "github.com/go-git/go-git/v6/plumbing.Reference" {
			return TypeMapping{Kind: MappingReference, GoType: "*plumbing.Reference", CType: "*C.char", CSharpType: "string", QualifiedName: "github.com/go-git/go-git/v6/plumbing.*Reference"}
		}
		if ptrQName == "github.com/go-git/go-git/v6/plumbing.Hash" || ptrQName == "github.com/go-git/go-git/v6/plumbing.ObjectID" {
			return TypeMapping{Kind: MappingHash, GoType: "*plumbing.Hash", CType: "*C.char", CSharpType: "string", QualifiedName: ptrQName}
		}

		if iterHandle, ok := knownPointerIterators[ptrQName]; ok {
			qname := pkg.Path() + ".*" + named.Obj().Name()
			return TypeMapping{Kind: MappingIterator, GoType: "*object." + named.Obj().Name(), CType: "C.longlong", CSharpType: "long", HandleType: iterHandle, QualifiedName: qname}
		}

		qname := pkg.Path() + ".*" + named.Obj().Name()

		if name, ok := dynamicHandles[qname]; ok {
			return TypeMapping{
				Kind:          MappingHandle,
				GoType:        t.String(),
				CType:         "C.longlong",
				CSharpType:    "long",
				HandleType:    name,
				QualifiedName: qname,
			}
		}
		if name, ok := dynamicOptions[qname]; ok {
			return TypeMapping{
				Kind:          MappingOptions,
				GoType:        t.String(),
				CType:         "C.longlong",
				CSharpType:    "long",
				HandleType:    name,
				QualifiedName: qname,
			}
		}

		return TypeMapping{Kind: MappingSkip, GoType: t.String()}

	case *types.Named:
		pkg := v.Obj().Pkg()
		name := v.Obj().Name()

		if pkg == nil {
			if name == "error" {
				return TypeMapping{Kind: MappingPrimitive, GoType: "error", CType: "*C.char", CSharpType: "IntPtr"}
			}
			return resolveTypeMapping(v.Underlying())
		}

		qname := pkg.Path() + "." + name

		switch qname {
		case "github.com/go-git/go-git/v6/plumbing.Hash",
			"github.com/go-git/go-git/v6/plumbing.ObjectID":
			return TypeMapping{Kind: MappingHash, GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string", QualifiedName: qname}
		case "github.com/go-git/go-git/v6/plumbing.ReferenceName":
			return TypeMapping{Kind: MappingReferenceName, GoType: "plumbing.ReferenceName", CType: "*C.char", CSharpType: "string", QualifiedName: qname}
		case "github.com/go-git/go-git/v6/plumbing.Revision":
			return TypeMapping{Kind: MappingRevision, GoType: "plumbing.Revision", CType: "*C.char", CSharpType: "string", QualifiedName: qname}
		case "github.com/go-git/go-git/v6/plumbing/transport.AuthMethod":
			return TypeMapping{Kind: MappingAuth, GoType: "transport.AuthMethod", CType: "C.longlong", CSharpType: "long", QualifiedName: qname}
		case "github.com/go-git/go-git/v6.Signer":
			return TypeMapping{Kind: MappingSigner, GoType: "git.Signer", CType: "C.longlong", CSharpType: "long", QualifiedName: qname}
		case "github.com/go-git/go-git/v6/plumbing/object.CommitIter":
			return TypeMapping{Kind: MappingIterator, GoType: "object.CommitIter", CType: "C.longlong", CSharpType: "long", HandleType: "CommitIter", QualifiedName: qname}
		case "github.com/go-git/go-git/v6/plumbing/storer.ReferenceIter":
			return TypeMapping{Kind: MappingIterator, GoType: "storer.ReferenceIter", CType: "C.longlong", CSharpType: "long", HandleType: "ReferenceIter", QualifiedName: qname}
		case "github.com/go-git/go-git/v6/plumbing/transport/server.Progress":
			return TypeMapping{Kind: MappingCallback, GoType: "sideband.Progress", CType: "C.longlong", CSharpType: "long", QualifiedName: qname}
		case "time.Time":
			return TypeMapping{Kind: MappingTime, GoType: "time.Time", CType: "C.longlong", CSharpType: "long", QualifiedName: qname}
		case "time.Duration":
			return TypeMapping{Kind: MappingDuration, GoType: "time.Duration", CType: "C.longlong", CSharpType: "long", QualifiedName: qname}
		}

		if dynamicEnums[qname] {
			return TypeMapping{Kind: MappingEnum, GoType: qname, CType: "C.int", CSharpType: "int", QualifiedName: qname}
		}

		if name, ok := dynamicHandles[qname]; ok {
			return TypeMapping{
				Kind:          MappingHandle,
				GoType:        qname,
				CType:         "C.longlong",
				CSharpType:    "long",
				HandleType:    name,
				QualifiedName: qname,
			}
		}

		return resolveTypeMapping(v.Underlying())

	case *types.Array:
		return TypeMapping{Kind: MappingSkip, GoType: t.String()}

	case *types.Slice:
		elem := v.Elem()
		if basic, ok := elem.(*types.Basic); ok && basic.Kind() == types.Byte {
			return TypeMapping{Kind: MappingByteSlice, GoType: "[]byte", CType: "*C.char", CSharpType: "byte[]"}
		}
		if basic, ok := elem.(*types.Basic); ok && basic.Kind() == types.String {
			return TypeMapping{Kind: MappingStringSlice, GoType: "[]string", CType: "*C.char", CSharpType: "string"}
		}
		elemToCheck := elem
		if alias, ok := elem.(*types.Alias); ok {
			elemToCheck = types.Unalias(alias)
		}
		if named, ok := elemToCheck.(*types.Named); ok {
			if named.Underlying() != nil {
				if basic, ok := named.Underlying().(*types.Basic); ok && basic.Kind() == types.String {
					elemQname := ""
					if named.Obj().Pkg() != nil {
						elemQname = named.Obj().Pkg().Path() + "." + named.Obj().Name()
					}
					return TypeMapping{Kind: MappingStringSlice, GoType: t.String(), CType: "*C.char", CSharpType: "string", QualifiedName: elemQname}
				}
			}
		}
		return TypeMapping{Kind: MappingSkip, GoType: t.String()}

	case *types.Interface:
		if v.NumMethods() == 1 && v.Method(0).Name() == "Error" {
			return TypeMapping{Kind: MappingPrimitive, GoType: "error", CType: "*C.char", CSharpType: "IntPtr"}
		}
		return TypeMapping{Kind: MappingSkip, GoType: t.String()}

	default:
		return TypeMapping{Kind: MappingSkip, GoType: t.String()}
	}
}

func resolveBasicMapping(b *types.Basic) TypeMapping {
	switch b.Kind() {
	case types.String:
		return TypeMapping{Kind: MappingString, GoType: "string", CType: "*C.char", CSharpType: "string"}
	case types.Bool:
		return TypeMapping{Kind: MappingBool, GoType: "bool", CType: "C.int", CSharpType: "int"}
	case types.Int, types.Int32:
		return TypeMapping{Kind: MappingPrimitive, GoType: b.Name(), CType: "C.int", CSharpType: "int"}
	case types.Int64:
		return TypeMapping{Kind: MappingPrimitive, GoType: "int64", CType: "C.longlong", CSharpType: "long"}
	case types.Uint, types.Uint32:
		return TypeMapping{Kind: MappingPrimitive, GoType: b.Name(), CType: "C.uint", CSharpType: "uint"}
	case types.Int8:
		return TypeMapping{Kind: MappingPrimitive, GoType: "int8", CType: "C.int", CSharpType: "int"}
	case types.Uint8:
		return TypeMapping{Kind: MappingPrimitive, GoType: "uint8", CType: "C.int", CSharpType: "int"}
	default:
		return TypeMapping{Kind: MappingSkip, GoType: b.Name()}
	}
}

func resolveMapping(goType string) TypeMapping {
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
	case "error":
		return TypeMapping{Kind: MappingPrimitive, GoType: "error", CType: "*C.char", CSharpType: "IntPtr"}
	case "plumbing.Hash":
		return TypeMapping{Kind: MappingHash, GoType: "plumbing.Hash", CType: "*C.char", CSharpType: "string"}
	case "plumbing.ReferenceName":
		return TypeMapping{Kind: MappingReferenceName, GoType: "plumbing.ReferenceName", CType: "*C.char", CSharpType: "string"}
	case "plumbing.Revision":
		return TypeMapping{Kind: MappingRevision, GoType: "plumbing.Revision", CType: "*C.char", CSharpType: "string"}
	case "*plumbing.Reference":
		return TypeMapping{Kind: MappingReference, GoType: "*plumbing.Reference", CType: "*C.char", CSharpType: "string"}
	case "transport.AuthMethod":
		return TypeMapping{Kind: MappingAuth, GoType: "transport.AuthMethod", CType: "C.longlong", CSharpType: "long"}
	case "git.Signer":
		return TypeMapping{Kind: MappingSigner, GoType: "git.Signer", CType: "C.longlong", CSharpType: "long"}
	case "object.CommitIter":
		return TypeMapping{Kind: MappingIterator, GoType: "object.CommitIter", CType: "C.longlong", CSharpType: "long", HandleType: "CommitIter"}
	case "storer.ReferenceIter":
		return TypeMapping{Kind: MappingIterator, GoType: "storer.ReferenceIter", CType: "C.longlong", CSharpType: "long", HandleType: "ReferenceIter"}
	case "*object.FileIter":
		return TypeMapping{Kind: MappingIterator, GoType: "*object.FileIter", CType: "C.longlong", CSharpType: "long", HandleType: "FileIter"}
	case "*object.TreeIter":
		return TypeMapping{Kind: MappingIterator, GoType: "*object.TreeIter", CType: "C.longlong", CSharpType: "long", HandleType: "TreeIter"}
	case "*object.BlobIter":
		return TypeMapping{Kind: MappingIterator, GoType: "*object.BlobIter", CType: "C.longlong", CSharpType: "long", HandleType: "BlobIter"}
	case "*object.TagIter":
		return TypeMapping{Kind: MappingIterator, GoType: "*object.TagIter", CType: "C.longlong", CSharpType: "long", HandleType: "TagIter"}
	}

	if name, ok := dynamicHandles["*"+goType]; ok {
		return TypeMapping{Kind: MappingHandle, GoType: goType, CType: "C.longlong", CSharpType: "long", HandleType: name}
	}
	if name, ok := dynamicOptions["*"+goType]; ok {
		return TypeMapping{Kind: MappingOptions, GoType: goType, CType: "C.longlong", CSharpType: "long", HandleType: name}
	}
	if dynamicEnums[goType] {
		return TypeMapping{Kind: MappingEnum, GoType: goType, CType: "C.int", CSharpType: "int"}
	}

	return TypeMapping{Kind: MappingSkip, GoType: goType}
}

func warnUnmappable(context, goType string) {
	fmt.Printf("WARNING: skipping %s — unmappable type %s\n", context, goType)
}
