package main

import (
	"fmt"
	"sort"
	"strings"
)

func generateOptionsGo(pkg *Package, outputDir string) error {
	var b strings.Builder
	b.WriteString(genHeader)
	b.WriteString("\n/*\n#include <stdlib.h>\n*/\nimport \"C\"\n")

	importSet := map[string]bool{
		"\"encoding/json\"": true,
		"\"time\"":          true,
	}
	for _, opts := range pkg.Options {
		addPackageImport(importSet, opts.ImportPath)
		for _, f := range opts.Fields {
			addMappingImports(importSet, f.Mapping)
		}
		if hasProxyOptions(opts.GoName) {
			addPackageImport(importSet, "github.com/go-git/go-git/v6/plumbing/transport")
		}
		if opts.GoName == "CreateTagOptions" {
			addPackageImport(importSet, "github.com/ProtonMail/go-crypto/openpgp")
		}
	}
	var imports []string
	for imp := range importSet {
		imports = append(imports, imp)
	}
	sort.Strings(imports)
	writeImports(&b, imports)
	b.WriteString("\n")

	for _, opts := range pkg.Options {
		goType := resolveOptionsGoType(opts)

		generateOptionsConstructor(&b, opts, goType)
		generateOptionsSetters(&b, opts, goType)
		generateOptionsSpecialSetters(&b, opts, goType)
		generateOptionsFree(&b, opts)
	}

	b.WriteString("var (\n")
	b.WriteString("\t_ = time.Now\n")
	b.WriteString("\t_ = json.Marshal\n")
	b.WriteString("\t_ object.Signature\n")
	b.WriteString("\t_ plumbing.Hash\n")
	b.WriteString("\t_ transport.AuthMethod\n")
	b.WriteString("\t_ *openpgp.Entity\n")
	b.WriteString(")\n")

	return writeGenFile(outputDir, "options_gen.go", b.String())
}

func resolveOptionsGoType(opts OptionsStruct) string {
	return importAlias(opts.ImportPath) + "." + opts.GoName
}

func generateOptionsConstructor(b *strings.Builder, opts OptionsStruct, goType string) {
	fmt.Fprintf(b, "//export %sNew\n", opts.CPrefix)
	fmt.Fprintf(b, "func %sNew(handleOut *C.longlong) {\n", opts.CPrefix)
	fmt.Fprintf(b, "\topts := &%s{}\n", goType)
	fmt.Fprintf(b, "\t*handleOut = C.longlong(storeHandle(opts))\n")
	fmt.Fprintf(b, "}\n\n")
}

func generateOptionsSetters(b *strings.Builder, opts OptionsStruct, goType string) {
	for _, f := range opts.Fields {
		setterName := opts.CPrefix + "Set" + f.GoName

		switch f.Mapping.Kind {
		case MappingString, MappingReferenceName, MappingBool, MappingPrimitive,
			MappingHash, MappingEnum, MappingAuth, MappingSigner,
			MappingTime, MappingDuration, MappingStringSlice:
		default:
			continue
		}

		fmt.Fprintf(b, "//export %s\n", setterName)

		switch f.Mapping.Kind {
		case MappingString, MappingReferenceName:
			fmt.Fprintf(b, "func %s(handle C.longlong, val *C.char) *C.char {\n", setterName)
			writeOptionsLoad(b, goType, opts.GoName)
			if f.Mapping.Kind == MappingReferenceName {
				fmt.Fprintf(b, "\topts.%s = plumbing.ReferenceName(C.GoString(val))\n", f.GoName)
			} else if f.Mapping.GoType == "*string" {
				fmt.Fprintf(b, "\ts := C.GoString(val)\n")
				fmt.Fprintf(b, "\topts.%s = &s\n", f.GoName)
			} else {
				fmt.Fprintf(b, "\topts.%s = C.GoString(val)\n", f.GoName)
			}
			fmt.Fprintf(b, "\treturn nil\n}\n\n")

		case MappingBool:
			fmt.Fprintf(b, "func %s(handle C.longlong, val C.int) *C.char {\n", setterName)
			writeOptionsLoad(b, goType, opts.GoName)
			fmt.Fprintf(b, "\topts.%s = val != 0\n", f.GoName)
			fmt.Fprintf(b, "\treturn nil\n}\n\n")

		case MappingPrimitive:
			cParamType := "C.int"
			goCast := "int"
			switch f.Mapping.GoType {
			case "int64":
				cParamType = "C.longlong"
				goCast = "int64"
			case "uint", "uint32":
				cParamType = "C.uint"
				goCast = f.Mapping.GoType
			case "int8":
				goCast = "int8"
			case "uint8":
				goCast = "uint8"
			}
			fmt.Fprintf(b, "func %s(handle C.longlong, val %s) *C.char {\n", setterName, cParamType)
			writeOptionsLoad(b, goType, opts.GoName)
			fmt.Fprintf(b, "\topts.%s = %s(val)\n", f.GoName, goCast)
			fmt.Fprintf(b, "\treturn nil\n}\n\n")

		case MappingHash:
			fmt.Fprintf(b, "func %s(handle C.longlong, val *C.char) *C.char {\n", setterName)
			writeOptionsLoad(b, goType, opts.GoName)
			fmt.Fprintf(b, "\th := plumbing.NewHash(C.GoString(val))\n")
			fmt.Fprintf(b, "\topts.%s = h\n", f.GoName)
			fmt.Fprintf(b, "\treturn nil\n}\n\n")

		case MappingEnum:
			fmt.Fprintf(b, "func %s(handle C.longlong, val C.int) *C.char {\n", setterName)
			writeOptionsLoad(b, goType, opts.GoName)
			fmt.Fprintf(b, "\topts.%s = %s(val)\n", f.GoName, resolveEnumCast(f.Mapping))
			fmt.Fprintf(b, "\treturn nil\n}\n\n")

		case MappingAuth:
			fmt.Fprintf(b, "func %s(handle C.longlong, authHandle C.longlong) *C.char {\n", setterName)
			writeOptionsLoad(b, goType, opts.GoName)
			fmt.Fprintf(b, "\tauth, ok := loadHandle[transport.AuthMethod](int64(authHandle))\n")
			fmt.Fprintf(b, "\tif !ok {\n\t\treturn C.CString(\"invalid auth handle\")\n\t}\n")
			fmt.Fprintf(b, "\topts.%s = auth\n", f.GoName)
			fmt.Fprintf(b, "\treturn nil\n}\n\n")

		case MappingSigner:
			fmt.Fprintf(b, "func %s(handle C.longlong, signerHandle C.longlong) *C.char {\n", setterName)
			writeOptionsLoad(b, goType, opts.GoName)
			fmt.Fprintf(b, "\tsigner, ok := loadHandle[git.Signer](int64(signerHandle))\n")
			fmt.Fprintf(b, "\tif !ok {\n\t\treturn C.CString(\"invalid signer handle\")\n\t}\n")
			fmt.Fprintf(b, "\topts.%s = signer\n", f.GoName)
			fmt.Fprintf(b, "\treturn nil\n}\n\n")

		case MappingTime:
			fmt.Fprintf(b, "func %s(handle C.longlong, unixSec C.longlong) *C.char {\n", setterName)
			writeOptionsLoad(b, goType, opts.GoName)
			if f.Mapping.GoType == "*time.Time" {
				fmt.Fprintf(b, "\tt := time.Unix(int64(unixSec), 0)\n")
				fmt.Fprintf(b, "\topts.%s = &t\n", f.GoName)
			} else {
				fmt.Fprintf(b, "\topts.%s = time.Unix(int64(unixSec), 0)\n", f.GoName)
			}
			fmt.Fprintf(b, "\treturn nil\n}\n\n")

		case MappingDuration:
			fmt.Fprintf(b, "func %s(handle C.longlong, nanos C.longlong) *C.char {\n", setterName)
			writeOptionsLoad(b, goType, opts.GoName)
			fmt.Fprintf(b, "\topts.%s = time.Duration(int64(nanos))\n", f.GoName)
			fmt.Fprintf(b, "\treturn nil\n}\n\n")

		case MappingStringSlice:
			fmt.Fprintf(b, "func %s(handle C.longlong, jsonVal *C.char) *C.char {\n", setterName)
			writeOptionsLoad(b, goType, opts.GoName)
			fmt.Fprintf(b, "\tvar ss []string\n")
			fmt.Fprintf(b, "\tif err := json.Unmarshal([]byte(C.GoString(jsonVal)), &ss); err != nil {\n")
			fmt.Fprintf(b, "\t\treturn toCError(err)\n\t}\n")
			if f.GoType != "[]string" && f.GoType != "" {
				sliceType := resolveSliceFieldType(f)
				fmt.Fprintf(b, "\tresult := make(%s, len(ss))\n", sliceType)
				elemType := sliceType[2:]
				fmt.Fprintf(b, "\tfor i, s := range ss {\n")
				fmt.Fprintf(b, "\t\tresult[i] = %s(s)\n", elemType)
				fmt.Fprintf(b, "\t}\n")
				fmt.Fprintf(b, "\topts.%s = result\n", f.GoName)
			} else {
				fmt.Fprintf(b, "\topts.%s = ss\n", f.GoName)
			}
			fmt.Fprintf(b, "\treturn nil\n}\n\n")
		}
	}
}

func generateOptionsSpecialSetters(b *strings.Builder, opts OptionsStruct, goType string) {
	if opts.GoName == "CommitOptions" {
		generateSignatureSetters(b, opts.CPrefix, goType, "Author")
		generateSignatureSetters(b, opts.CPrefix, goType, "Committer")
		generateHashSliceSetter(b, opts.CPrefix, goType, "Parents")
	}
	if opts.GoName == "CreateTagOptions" {
		generateSignatureSetters(b, opts.CPrefix, goType, "Tagger")
		generateSignKeyHandleSetter(b, opts.CPrefix, goType)
	}
	if opts.GoName == "RestoreOptions" {
		fmt.Fprintf(b, "//export %sAddFile\n", opts.CPrefix)
		fmt.Fprintf(b, "func %sAddFile(handle C.longlong, path *C.char) *C.char {\n", opts.CPrefix)
		writeOptionsLoad(b, goType, opts.GoName)
		fmt.Fprintf(b, "\topts.Files = append(opts.Files, C.GoString(path))\n")
		fmt.Fprintf(b, "\treturn nil\n}\n\n")
	}
	if hasProxyOptions(opts.GoName) {
		generateProxySetter(b, opts.CPrefix, goType)
	}
	if opts.GoName == "PushOptions" {
		generateForceWithLeaseSetter(b, opts.CPrefix, goType)
	}
}

func hasProxyOptions(name string) bool {
	switch name {
	case "CloneOptions", "FetchOptions", "ListOptions", "PullOptions", "PushOptions":
		return true
	}
	return false
}

func generateProxySetter(b *strings.Builder, cPrefix, goType string) {
	fmt.Fprintf(b, "//export %sSetProxy\n", cPrefix)
	fmt.Fprintf(b, "func %sSetProxy(handle C.longlong, url *C.char, username *C.char, password *C.char) *C.char {\n", cPrefix)
	writeOptionsLoad(b, goType, goType[strings.LastIndex(goType, ".")+1:])
	fmt.Fprintf(b, "\topts.ProxyOptions = transport.ProxyOptions{\n")
	fmt.Fprintf(b, "\t\tURL:      C.GoString(url),\n")
	fmt.Fprintf(b, "\t\tUsername: C.GoString(username),\n")
	fmt.Fprintf(b, "\t\tPassword: C.GoString(password),\n")
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\treturn nil\n}\n\n")
}

func generateHashSliceSetter(b *strings.Builder, cPrefix, goType, fieldName string) {
	fmt.Fprintf(b, "//export %sSet%s\n", cPrefix, fieldName)
	fmt.Fprintf(b, "func %sSet%s(handle C.longlong, jsonHashes *C.char) *C.char {\n", cPrefix, fieldName)
	writeOptionsLoad(b, goType, goType[strings.LastIndex(goType, ".")+1:])
	fmt.Fprintf(b, "\tvar hexes []string\n")
	fmt.Fprintf(b, "\tif err := json.Unmarshal([]byte(C.GoString(jsonHashes)), &hexes); err != nil {\n")
	fmt.Fprintf(b, "\t\treturn toCError(err)\n\t}\n")
	fmt.Fprintf(b, "\thashes := make([]plumbing.Hash, len(hexes))\n")
	fmt.Fprintf(b, "\tfor i, h := range hexes {\n")
	fmt.Fprintf(b, "\t\thashes[i] = plumbing.NewHash(h)\n")
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\topts.%s = hashes\n", fieldName)
	fmt.Fprintf(b, "\treturn nil\n}\n\n")
}

func generateForceWithLeaseSetter(b *strings.Builder, cPrefix, goType string) {
	fmt.Fprintf(b, "//export %sSetForceWithLease\n", cPrefix)
	fmt.Fprintf(b, "func %sSetForceWithLease(handle C.longlong, refName *C.char, hash *C.char) *C.char {\n", cPrefix)
	writeOptionsLoad(b, goType, goType[strings.LastIndex(goType, ".")+1:])
	fmt.Fprintf(b, "\topts.ForceWithLease = &git.ForceWithLease{\n")
	fmt.Fprintf(b, "\t\tRefName: plumbing.ReferenceName(C.GoString(refName)),\n")
	fmt.Fprintf(b, "\t\tHash:    plumbing.NewHash(C.GoString(hash)),\n")
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\treturn nil\n}\n\n")
}

func generateSignKeyHandleSetter(b *strings.Builder, cPrefix, goType string) {
	fmt.Fprintf(b, "//export %sSetSignKey\n", cPrefix)
	fmt.Fprintf(b, "func %sSetSignKey(handle C.longlong, keyHandle C.longlong) *C.char {\n", cPrefix)
	writeOptionsLoad(b, goType, goType[strings.LastIndex(goType, ".")+1:])
	fmt.Fprintf(b, "\tentity, ok := loadHandle[*openpgp.Entity](int64(keyHandle))\n")
	fmt.Fprintf(b, "\tif !ok {\n\t\treturn C.CString(\"invalid signing key handle\")\n\t}\n")
	fmt.Fprintf(b, "\topts.SignKey = entity\n")
	fmt.Fprintf(b, "\treturn nil\n}\n\n")
}

func generateOptionsFree(b *strings.Builder, opts OptionsStruct) {
	fmt.Fprintf(b, "//export %sFree\n", opts.CPrefix)
	fmt.Fprintf(b, "func %sFree(handle C.longlong) {\n", opts.CPrefix)
	fmt.Fprintf(b, "\tremoveHandle(int64(handle))\n")
	fmt.Fprintf(b, "}\n\n")
}

func writeOptionsLoad(b *strings.Builder, goType, goName string) {
	fmt.Fprintf(b, "\topts, ok := loadHandle[*%s](int64(handle))\n", goType)
	fmt.Fprintf(b, "\tif !ok {\n\t\treturn C.CString(\"invalid %s handle\")\n\t}\n", goName)
}

func resolveSliceFieldType(f OptionsField) string {
	if f.Mapping.QualifiedName == "" {
		return "[]string"
	}
	lastDot := strings.LastIndex(f.Mapping.QualifiedName, ".")
	if lastDot < 0 {
		return "[]string"
	}
	pkgPath := f.Mapping.QualifiedName[:lastDot]
	typeName := f.Mapping.QualifiedName[lastDot+1:]
	return "[]" + importAlias(pkgPath) + "." + typeName
}

func resolveEnumCast(m TypeMapping) string {
	qname := m.QualifiedName
	if qname == "" {
		qname = m.GoType
	}

	lastDot := strings.LastIndex(qname, ".")
	if lastDot < 0 {
		return qname
	}
	pkgPath := qname[:lastDot]
	typeName := qname[lastDot+1:]
	return importAlias(pkgPath) + "." + typeName
}
