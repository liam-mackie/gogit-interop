package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const sharedPkg = "cmd/shared"

func generateGo(pkg *Package, outputDir string) error {
	if err := generateAuthGo(outputDir); err != nil {
		return fmt.Errorf("generating auth: %w", err)
	}
	if err := generateSigningGo(outputDir); err != nil {
		return fmt.Errorf("generating signing: %w", err)
	}
	if err := generateOptionsGo(pkg, outputDir); err != nil {
		return fmt.Errorf("generating options: %w", err)
	}
	if err := generateRepositoryGo(pkg, outputDir); err != nil {
		return fmt.Errorf("generating repository: %w", err)
	}
	if err := generateWorktreeGo(pkg, outputDir); err != nil {
		return fmt.Errorf("generating worktree: %w", err)
	}
	if err := generateRemoteGo(pkg, outputDir); err != nil {
		return fmt.Errorf("generating remote: %w", err)
	}
	if err := generateSubmoduleGo(pkg, outputDir); err != nil {
		return fmt.Errorf("generating submodule: %w", err)
	}
	if err := generateIteratorsGo(outputDir); err != nil {
		return fmt.Errorf("generating iterators: %w", err)
	}
	return nil
}

func writeGenFile(outputDir, filename, content string) error {
	path := filepath.Join(outputDir, filename)
	return os.WriteFile(path, []byte(content), 0644)
}

func generateAuthGo(outputDir string) error {
	content := `package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
)

//export GitAuthNewBasicHTTP
func GitAuthNewBasicHTTP(username *C.char, password *C.char, handleOut *C.longlong) *C.char {
	auth := &http.BasicAuth{
		Username: C.GoString(username),
		Password: C.GoString(password),
	}
	*handleOut = C.longlong(storeHandle(auth))
	return nil
}

//export GitAuthNewTokenHTTP
func GitAuthNewTokenHTTP(token *C.char, handleOut *C.longlong) *C.char {
	auth := &http.TokenAuth{
		Token: C.GoString(token),
	}
	*handleOut = C.longlong(storeHandle(auth))
	return nil
}

//export GitAuthNewSSHKeyFromFile
func GitAuthNewSSHKeyFromFile(user *C.char, pemFile *C.char, password *C.char, handleOut *C.longlong) *C.char {
	auth, err := ssh.NewPublicKeysFromFile(C.GoString(user), C.GoString(pemFile), C.GoString(password))
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(auth))
	return nil
}

//export GitAuthNewSSHKey
func GitAuthNewSSHKey(user *C.char, pem *C.char, password *C.char, handleOut *C.longlong) *C.char {
	auth, err := ssh.NewPublicKeys(C.GoString(user), []byte(C.GoString(pem)), C.GoString(password))
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(auth))
	return nil
}

//export GitAuthNewSSHAgent
func GitAuthNewSSHAgent(user *C.char, handleOut *C.longlong) *C.char {
	auth, err := ssh.NewSSHAgentAuth(C.GoString(user))
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(auth))
	return nil
}

//export GitAuthNewSSHPassword
func GitAuthNewSSHPassword(user *C.char, password *C.char, handleOut *C.longlong) *C.char {
	auth := &ssh.Password{
		User:     C.GoString(user),
		Password: C.GoString(password),
	}
	*handleOut = C.longlong(storeHandle(auth))
	return nil
}

//export GitAuthFree
func GitAuthFree(handle C.longlong) {
	removeHandle(int64(handle))
}
`
	return writeGenFile(outputDir, "auth_gen.go", content)
}

func generateSigningGo(outputDir string) error {
	content := `package main

/*
#include <stdlib.h>
#include "callbacks.h"
*/
import "C"
import (
	"bytes"
	"errors"
	"io"
	"unsafe"

	"github.com/ProtonMail/go-crypto/openpgp"
)

//export GitSignerNewPGP
func GitSignerNewPGP(armoredKey *C.char, passphrase *C.char, handleOut *C.longlong) *C.char {
	keyRing, err := openpgp.ReadArmoredKeyRing(bytes.NewReader([]byte(C.GoString(armoredKey))))
	if err != nil {
		return toCError(err)
	}
	if len(keyRing) == 0 {
		return C.CString("no keys found in armored key")
	}
	entity := keyRing[0]

	pp := C.GoString(passphrase)
	if pp != "" {
		if entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
			if err := entity.PrivateKey.Decrypt([]byte(pp)); err != nil {
				return toCError(err)
			}
		}
		for _, sub := range entity.Subkeys {
			if sub.PrivateKey != nil && sub.PrivateKey.Encrypted {
				_ = sub.PrivateKey.Decrypt([]byte(pp))
			}
		}
	}

	signer := &pgpSigner{entity: entity}
	*handleOut = C.longlong(storeHandle(signer))
	return nil
}

type pgpSigner struct {
	entity *openpgp.Entity
}

func (s *pgpSigner) Sign(message io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	if err := openpgp.ArmoredDetachSign(&buf, s.entity, message, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//export GitSignerNewCallback
func GitSignerNewCallback(fn C.GitSignFunc, userData unsafe.Pointer, handleOut *C.longlong) *C.char {
	signer := &callbackSigner{fn: fn, userData: userData}
	*handleOut = C.longlong(storeHandle(signer))
	return nil
}

type callbackSigner struct {
	fn       C.GitSignFunc
	userData unsafe.Pointer
}

func (s *callbackSigner) Sign(message io.Reader) ([]byte, error) {
	data, err := io.ReadAll(message)
	if err != nil {
		return nil, err
	}

	var sigOut *C.char
	var sigLen C.int
	errStr := C.callSignFunc(s.fn, (*C.char)(unsafe.Pointer(&data[0])), C.int(len(data)), &sigOut, &sigLen, s.userData)
	if errStr != nil {
		defer C.free(unsafe.Pointer(errStr))
		return nil, errors.New(C.GoString(errStr))
	}

	sig := C.GoBytes(unsafe.Pointer(sigOut), sigLen)
	C.free(unsafe.Pointer(sigOut))
	return sig, nil
}

//export GitSigningKeyNewPGP
func GitSigningKeyNewPGP(armoredKey *C.char, passphrase *C.char, handleOut *C.longlong) *C.char {
	keyRing, err := openpgp.ReadArmoredKeyRing(bytes.NewReader([]byte(C.GoString(armoredKey))))
	if err != nil {
		return toCError(err)
	}
	if len(keyRing) == 0 {
		return C.CString("no keys found in armored key")
	}
	entity := keyRing[0]

	pp := C.GoString(passphrase)
	if pp != "" {
		if entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
			if err := entity.PrivateKey.Decrypt([]byte(pp)); err != nil {
				return toCError(err)
			}
		}
	}

	*handleOut = C.longlong(storeHandle(entity))
	return nil
}

//export GitSignerFree
func GitSignerFree(handle C.longlong) {
	removeHandle(int64(handle))
}

//export GitSigningKeyFree
func GitSigningKeyFree(handle C.longlong) {
	removeHandle(int64(handle))
}
`
	return writeGenFile(outputDir, "signing_gen.go", content)
}

func generateOptionsGo(pkg *Package, outputDir string) error {
	var b strings.Builder
	b.WriteString(`package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport"
)

`)

	for _, opts := range pkg.Options {
		goType := "git." + opts.GoName

		// Constructor
		fmt.Fprintf(&b, "//export %sNew\n", opts.CPrefix)
		fmt.Fprintf(&b, "func %sNew(handleOut *C.longlong) {\n", opts.CPrefix)
		fmt.Fprintf(&b, "\topts := &%s{}\n", goType)
		fmt.Fprintf(&b, "\t*handleOut = C.longlong(storeHandle(opts))\n")
		fmt.Fprintf(&b, "}\n\n")

		// Setters
		for _, f := range opts.Fields {
			setterName := opts.CPrefix + "Set" + f.GoName
			fmt.Fprintf(&b, "//export %s\n", setterName)

			switch f.Mapping.Kind {
			case MappingString, MappingReferenceName:
				fmt.Fprintf(&b, "func %s(handle C.longlong, val *C.char) *C.char {\n", setterName)
				fmt.Fprintf(&b, "\topts, ok := loadHandle[*%s](int64(handle))\n", goType)
				fmt.Fprintf(&b, "\tif !ok {\n\t\treturn C.CString(\"invalid %s handle\")\n\t}\n", opts.GoName)
				if f.GoType == "plumbing.ReferenceName" {
					fmt.Fprintf(&b, "\topts.%s = plumbing.ReferenceName(C.GoString(val))\n", f.GoName)
				} else {
					fmt.Fprintf(&b, "\topts.%s = C.GoString(val)\n", f.GoName)
				}
				fmt.Fprintf(&b, "\treturn nil\n}\n\n")

			case MappingBool:
				fmt.Fprintf(&b, "func %s(handle C.longlong, val C.int) *C.char {\n", setterName)
				fmt.Fprintf(&b, "\topts, ok := loadHandle[*%s](int64(handle))\n", goType)
				fmt.Fprintf(&b, "\tif !ok {\n\t\treturn C.CString(\"invalid %s handle\")\n\t}\n", opts.GoName)
				fmt.Fprintf(&b, "\topts.%s = val != 0\n", f.GoName)
				fmt.Fprintf(&b, "\treturn nil\n}\n\n")

			case MappingPrimitive:
				fmt.Fprintf(&b, "func %s(handle C.longlong, val C.int) *C.char {\n", setterName)
				fmt.Fprintf(&b, "\topts, ok := loadHandle[*%s](int64(handle))\n", goType)
				fmt.Fprintf(&b, "\tif !ok {\n\t\treturn C.CString(\"invalid %s handle\")\n\t}\n", opts.GoName)
				fmt.Fprintf(&b, "\topts.%s = int(val)\n", f.GoName)
				fmt.Fprintf(&b, "\treturn nil\n}\n\n")

			case MappingHash:
				fmt.Fprintf(&b, "func %s(handle C.longlong, val *C.char) *C.char {\n", setterName)
				fmt.Fprintf(&b, "\topts, ok := loadHandle[*%s](int64(handle))\n", goType)
				fmt.Fprintf(&b, "\tif !ok {\n\t\treturn C.CString(\"invalid %s handle\")\n\t}\n", opts.GoName)
				fmt.Fprintf(&b, "\th := plumbing.NewHash(C.GoString(val))\n")
				fmt.Fprintf(&b, "\topts.%s = h\n", f.GoName)
				fmt.Fprintf(&b, "\treturn nil\n}\n\n")

			case MappingEnum:
				cType := "C.int"
				fmt.Fprintf(&b, "func %s(handle C.longlong, val %s) *C.char {\n", setterName, cType)
				fmt.Fprintf(&b, "\topts, ok := loadHandle[*%s](int64(handle))\n", goType)
				fmt.Fprintf(&b, "\tif !ok {\n\t\treturn C.CString(\"invalid %s handle\")\n\t}\n", opts.GoName)

				switch f.GoType {
				case "plumbing.TagMode":
					fmt.Fprintf(&b, "\topts.%s = plumbing.TagMode(val)\n", f.GoName)
				case "git.ResetMode":
					fmt.Fprintf(&b, "\topts.%s = git.ResetMode(val)\n", f.GoName)
				case "git.LogOrder":
					fmt.Fprintf(&b, "\topts.%s = git.LogOrder(val)\n", f.GoName)
				case "git.MergeStrategy":
					fmt.Fprintf(&b, "\topts.%s = git.MergeStrategy(val)\n", f.GoName)
				default:
					fmt.Fprintf(&b, "\topts.%s = %s(val)\n", f.GoName, f.GoType)
				}
				fmt.Fprintf(&b, "\treturn nil\n}\n\n")

			case MappingAuth:
				fmt.Fprintf(&b, "func %s(handle C.longlong, authHandle C.longlong) *C.char {\n", setterName)
				fmt.Fprintf(&b, "\topts, ok := loadHandle[*%s](int64(handle))\n", goType)
				fmt.Fprintf(&b, "\tif !ok {\n\t\treturn C.CString(\"invalid %s handle\")\n\t}\n", opts.GoName)
				fmt.Fprintf(&b, "\tauth, ok := loadHandle[transport.AuthMethod](int64(authHandle))\n")
				fmt.Fprintf(&b, "\tif !ok {\n\t\treturn C.CString(\"invalid auth handle\")\n\t}\n")
				fmt.Fprintf(&b, "\topts.%s = auth\n", f.GoName)
				fmt.Fprintf(&b, "\treturn nil\n}\n\n")

			case MappingSigner:
				fmt.Fprintf(&b, "func %s(handle C.longlong, signerHandle C.longlong) *C.char {\n", setterName)
				fmt.Fprintf(&b, "\topts, ok := loadHandle[*%s](int64(handle))\n", goType)
				fmt.Fprintf(&b, "\tif !ok {\n\t\treturn C.CString(\"invalid %s handle\")\n\t}\n", opts.GoName)
				fmt.Fprintf(&b, "\tsigner, ok := loadHandle[git.Signer](int64(signerHandle))\n")
				fmt.Fprintf(&b, "\tif !ok {\n\t\treturn C.CString(\"invalid signer handle\")\n\t}\n")
				fmt.Fprintf(&b, "\topts.%s = signer\n", f.GoName)
				fmt.Fprintf(&b, "\treturn nil\n}\n\n")
			}
		}

		// Special setters for CommitOptions (Author/Committer as Signature)
		if opts.GoName == "CommitOptions" {
			generateSignatureSetters(&b, opts.CPrefix, goType, "Author")
			generateSignatureSetters(&b, opts.CPrefix, goType, "Committer")
		}

		if opts.GoName == "CreateTagOptions" {
			generateSignatureSetters(&b, opts.CPrefix, goType, "Tagger")
		}

		// RestoreOptions files setter
		if opts.GoName == "RestoreOptions" {
			fmt.Fprintf(&b, "//export %sAddFile\n", opts.CPrefix)
			fmt.Fprintf(&b, "func %sAddFile(handle C.longlong, path *C.char) *C.char {\n", opts.CPrefix)
			fmt.Fprintf(&b, "\topts, ok := loadHandle[*%s](int64(handle))\n", goType)
			fmt.Fprintf(&b, "\tif !ok {\n\t\treturn C.CString(\"invalid %s handle\")\n\t}\n", opts.GoName)
			fmt.Fprintf(&b, "\topts.Files = append(opts.Files, C.GoString(path))\n")
			fmt.Fprintf(&b, "\treturn nil\n}\n\n")
		}

		// Free
		fmt.Fprintf(&b, "//export %sFree\n", opts.CPrefix)
		fmt.Fprintf(&b, "func %sFree(handle C.longlong) {\n", opts.CPrefix)
		fmt.Fprintf(&b, "\tremoveHandle(int64(handle))\n")
		fmt.Fprintf(&b, "}\n\n")
	}

	// Suppress unused import warnings
	b.WriteString("var (\n")
	b.WriteString("\t_ = time.Now\n")
	b.WriteString("\t_ object.Signature\n")
	b.WriteString(")\n")

	return writeGenFile(outputDir, "options_gen.go", b.String())
}

func generateSignatureSetters(b *strings.Builder, cPrefix, goType, fieldName string) {
	// The CommitOptions uses Author/Committer as *object.Signature
	// We provide Name+Email setter that creates the Signature
	fmt.Fprintf(b, "//export %sSet%sNameEmail\n", cPrefix, fieldName)
	fmt.Fprintf(b, "func %sSet%sNameEmail(handle C.longlong, name *C.char, email *C.char) *C.char {\n", cPrefix, fieldName)
	fmt.Fprintf(b, "\topts, ok := loadHandle[*%s](int64(handle))\n", goType)
	fmt.Fprintf(b, "\tif !ok {\n\t\treturn C.CString(\"invalid options handle\")\n\t}\n")
	fmt.Fprintf(b, "\topts.%s = &object.Signature{\n", fieldName)
	fmt.Fprintf(b, "\t\tName:  C.GoString(name),\n")
	fmt.Fprintf(b, "\t\tEmail: C.GoString(email),\n")
	fmt.Fprintf(b, "\t\tWhen:  time.Now(),\n")
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\treturn nil\n}\n\n")
}

func generateRepositoryGo(pkg *Package, outputDir string) error {
	var repoType *HandleType
	for i := range pkg.Types {
		if pkg.Types[i].GoName == "Repository" {
			repoType = &pkg.Types[i]
			break
		}
	}
	if repoType == nil {
		return fmt.Errorf("repository type not found")
	}

	var b strings.Builder
	b.WriteString(`package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/transport"
)

`)

	// Top-level functions
	for _, fn := range pkg.Functions {
		generateTopLevelFunction(&b, fn)
	}

	// Repository methods
	for _, m := range repoType.Methods {
		generateRepositoryMethod(&b, m)
	}

	// Remotes list (returns JSON array of remote names)
	b.WriteString(`//export GitRepositoryRemotes
func GitRepositoryRemotes(repoHandle C.longlong, jsonOut **C.char) *C.char {
	repo, ok := loadHandle[*git.Repository](int64(repoHandle))
	if !ok {
		return C.CString("invalid repository handle")
	}
	remotes, err := repo.Remotes()
	if err != nil {
		return toCError(err)
	}
	names := make([]string, len(remotes))
	for i, r := range remotes {
		names[i] = r.Config().Name
	}
	data, err := json.Marshal(names)
	if err != nil {
		return toCError(err)
	}
	*jsonOut = C.CString(string(data))
	return nil
}

//export GitRepositoryFree
func GitRepositoryFree(repoHandle C.longlong) {
	removeHandle(int64(repoHandle))
}

`)

	// Suppress unused imports
	b.WriteString("var (\n")
	b.WriteString("\t_ = json.Marshal\n")
	b.WriteString("\t_ config.RemoteConfig\n")
	b.WriteString("\t_ transport.AuthMethod\n")
	b.WriteString(")\n")

	return writeGenFile(outputDir, "repository_gen.go", b.String())
}

func generateTopLevelFunction(b *strings.Builder, fn Function) {
	fmt.Fprintf(b, "//export %s\n", fn.CName)

	// Build function signature
	var params []string
	for _, p := range fn.Params {
		switch p.Mapping.Kind {
		case MappingString, MappingReferenceName:
			params = append(params, fmt.Sprintf("%s *C.char", p.CName))
		case MappingBool:
			params = append(params, fmt.Sprintf("%s C.int", p.CName))
		case MappingOptions, MappingHandle:
			params = append(params, fmt.Sprintf("%s C.longlong", p.CName))
		case MappingPrimitive:
			params = append(params, fmt.Sprintf("%s C.int", p.CName))
		}
	}

	hasHandleReturn := false
	for _, r := range fn.Returns {
		if !r.IsError && r.Mapping.Kind == MappingHandle {
			params = append(params, "handleOut *C.longlong")
			hasHandleReturn = true
		}
	}

	fmt.Fprintf(b, "func %s(%s) *C.char {\n", fn.CName, strings.Join(params, ", "))

	// Build Go call
	switch fn.GoName {
	case "PlainInit":
		fmt.Fprintf(b, "\trepo, err := git.PlainInit(C.GoString(path), isBare != 0)\n")
		fmt.Fprintf(b, "\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
		fmt.Fprintf(b, "\t*handleOut = C.longlong(storeHandle(repo))\n")
	case "PlainOpen":
		fmt.Fprintf(b, "\trepo, err := git.PlainOpen(C.GoString(path))\n")
		fmt.Fprintf(b, "\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
		fmt.Fprintf(b, "\t*handleOut = C.longlong(storeHandle(repo))\n")
	case "PlainOpenWithOptions":
		fmt.Fprintf(b, "\topts, ok := loadHandle[*git.PlainOpenOptions](int64(optsHandle))\n")
		fmt.Fprintf(b, "\tif !ok {\n\t\treturn C.CString(\"invalid PlainOpenOptions handle\")\n\t}\n")
		fmt.Fprintf(b, "\trepo, err := git.PlainOpenWithOptions(C.GoString(path), opts)\n")
		fmt.Fprintf(b, "\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
		fmt.Fprintf(b, "\t*handleOut = C.longlong(storeHandle(repo))\n")
	case "PlainCloneWithOptions":
		fmt.Fprintf(b, "\topts, ok := loadHandle[*git.CloneOptions](int64(optsHandle))\n")
		fmt.Fprintf(b, "\tif !ok {\n\t\treturn C.CString(\"invalid CloneOptions handle\")\n\t}\n")
		fmt.Fprintf(b, "\trepo, err := git.PlainClone(C.GoString(path), opts)\n")
		fmt.Fprintf(b, "\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
		fmt.Fprintf(b, "\t*handleOut = C.longlong(storeHandle(repo))\n")
	}

	if hasHandleReturn {
		fmt.Fprintf(b, "\treturn nil\n")
	}
	fmt.Fprintf(b, "}\n\n")

}

func generateRepositoryMethod(b *strings.Builder, m Method) {
	switch m.GoName {
	case "Head":
		generateRepoHead(b, m)
	case "Worktree":
		generateRepoWorktree(b, m)
	case "Fetch":
		generateRepoFetch(b, m)
	case "Push":
		generateRepoPush(b, m)
	case "Log":
		generateRepoLog(b, m)
	case "Tags", "Branches", "Notes", "References":
		generateRepoIterMethod(b, m)
	case "Reference":
		generateRepoReference(b, m)
	case "ResolveRevision":
		generateRepoResolveRevision(b, m)
	case "CreateRemote":
		generateRepoCreateRemote(b, m)
	case "Remote":
		generateRepoGetRemote(b, m)
	case "DeleteRemote":
		generateRepoDeleteRemote(b, m)
	case "CreateBranch":
		generateRepoCreateBranch(b, m)
	case "DeleteBranch":
		generateRepoDeleteBranch(b, m)
	case "CreateTag":
		generateRepoCreateTag(b, m)
	case "Tag":
		generateRepoGetTag(b, m)
	case "DeleteTag":
		generateRepoDeleteTag(b, m)
	case "CommitObject":
		generateRepoCommitObject(b, m)
	case "Merge":
		generateRepoMerge(b, m)
	}
}

func repoLoadPreamble(b *strings.Builder) {
	b.WriteString("\trepo, ok := loadHandle[*git.Repository](int64(repoHandle))\n")
	b.WriteString("\tif !ok {\n\t\treturn C.CString(\"invalid repository handle\")\n\t}\n")
}

func generateRepoHead(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, refNameOut **C.char, hashOut **C.char) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\tref, err := repo.Head()\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*refNameOut = C.CString(string(ref.Name()))\n")
	b.WriteString("\t*hashOut = C.CString(ref.Hash().String())\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRepoWorktree(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, wtHandleOut *C.longlong) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\twt, err := repo.Worktree()\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*wtHandleOut = C.longlong(storeHandle(wt))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRepoFetch(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, optsHandle C.longlong) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\topts, ok := loadHandle[*git.FetchOptions](int64(optsHandle))\n")
	b.WriteString("\tif !ok {\n\t\treturn C.CString(\"invalid FetchOptions handle\")\n\t}\n")
	b.WriteString("\treturn toCError(repo.Fetch(opts))\n}\n\n")
}

func generateRepoPush(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, optsHandle C.longlong) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\topts, ok := loadHandle[*git.PushOptions](int64(optsHandle))\n")
	b.WriteString("\tif !ok {\n\t\treturn C.CString(\"invalid PushOptions handle\")\n\t}\n")
	b.WriteString("\treturn toCError(repo.Push(opts))\n}\n\n")
}

func generateRepoLog(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, optsHandle C.longlong, iterOut *C.longlong) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\tvar opts *git.LogOptions\n")
	b.WriteString("\tif int64(optsHandle) != 0 {\n")
	b.WriteString("\t\tvar ok bool\n")
	b.WriteString("\t\topts, ok = loadHandle[*git.LogOptions](int64(optsHandle))\n")
	b.WriteString("\t\tif !ok {\n\t\t\treturn C.CString(\"invalid LogOptions handle\")\n\t\t}\n")
	b.WriteString("\t} else {\n")
	b.WriteString("\t\topts = &git.LogOptions{}\n")
	b.WriteString("\t}\n")
	b.WriteString("\titer, err := repo.Log(opts)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*iterOut = C.longlong(storeHandle(iter))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRepoIterMethod(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, iterOut *C.longlong) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	fmt.Fprintf(b, "\titer, err := repo.%s()\n", m.GoName)
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*iterOut = C.longlong(storeHandle(iter))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRepoReference(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char, resolved C.int, refNameOut **C.char, hashOut **C.char) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\tref, err := repo.Reference(plumbing.ReferenceName(C.GoString(name)), resolved != 0)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*refNameOut = C.CString(string(ref.Name()))\n")
	b.WriteString("\t*hashOut = C.CString(ref.Hash().String())\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRepoResolveRevision(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, rev *C.char, hashOut **C.char) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\thash, err := repo.ResolveRevision(plumbing.Revision(C.GoString(rev)))\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*hashOut = C.CString(hash.String())\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRepoCreateRemote(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char, url *C.char, handleOut *C.longlong) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\tremote, err := repo.CreateRemote(&config.RemoteConfig{\n")
	b.WriteString("\t\tName: C.GoString(name),\n")
	b.WriteString("\t\tURLs: []string{C.GoString(url)},\n")
	b.WriteString("\t})\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*handleOut = C.longlong(storeHandle(remote))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRepoGetRemote(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char, handleOut *C.longlong) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\tremote, err := repo.Remote(C.GoString(name))\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*handleOut = C.longlong(storeHandle(remote))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRepoDeleteRemote(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\treturn toCError(repo.DeleteRemote(C.GoString(name)))\n}\n\n")
}

func generateRepoCreateBranch(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char, hash *C.char) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\tbranch := &config.Branch{\n")
	b.WriteString("\t\tName: C.GoString(name),\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn toCError(repo.CreateBranch(branch))\n}\n\n")
}

func generateRepoDeleteBranch(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\treturn toCError(repo.DeleteBranch(C.GoString(name)))\n}\n\n")
}

func generateRepoCreateTag(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char, hash *C.char, optsHandle C.longlong, refNameOut **C.char, hashOut **C.char) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\th := plumbing.NewHash(C.GoString(hash))\n")
	b.WriteString("\tvar opts *git.CreateTagOptions\n")
	b.WriteString("\tif int64(optsHandle) != 0 {\n")
	b.WriteString("\t\tvar ok bool\n")
	b.WriteString("\t\topts, ok = loadHandle[*git.CreateTagOptions](int64(optsHandle))\n")
	b.WriteString("\t\tif !ok {\n\t\t\treturn C.CString(\"invalid CreateTagOptions handle\")\n\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("\tref, err := repo.CreateTag(C.GoString(name), h, opts)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*refNameOut = C.CString(string(ref.Name()))\n")
	b.WriteString("\t*hashOut = C.CString(ref.Hash().String())\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRepoGetTag(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char, refNameOut **C.char, hashOut **C.char) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\tref, err := repo.Tag(C.GoString(name))\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*refNameOut = C.CString(string(ref.Name()))\n")
	b.WriteString("\t*hashOut = C.CString(ref.Hash().String())\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRepoDeleteTag(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, name *C.char) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\treturn toCError(repo.DeleteTag(C.GoString(name)))\n}\n\n")
}

func generateRepoCommitObject(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, hash *C.char, commitHashOut **C.char, msgOut **C.char, authorNameOut **C.char, authorEmailOut **C.char, tsOut *C.longlong) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\th := plumbing.NewHash(C.GoString(hash))\n")
	b.WriteString("\tcommit, err := repo.CommitObject(h)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*commitHashOut = C.CString(commit.Hash.String())\n")
	b.WriteString("\t*msgOut = C.CString(commit.Message)\n")
	b.WriteString("\t*authorNameOut = C.CString(commit.Author.Name)\n")
	b.WriteString("\t*authorEmailOut = C.CString(commit.Author.Email)\n")
	b.WriteString("\t*tsOut = C.longlong(commit.Author.When.Unix())\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRepoMerge(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(repoHandle C.longlong, refName *C.char, hash *C.char, optsHandle C.longlong) *C.char {\n", m.CName)
	repoLoadPreamble(b)
	b.WriteString("\th := plumbing.NewHash(C.GoString(hash))\n")
	b.WriteString("\tref := plumbing.NewHashReference(plumbing.ReferenceName(C.GoString(refName)), h)\n")
	b.WriteString("\tvar opts git.MergeOptions\n")
	b.WriteString("\tif int64(optsHandle) != 0 {\n")
	b.WriteString("\t\toptsPtr, ok := loadHandle[*git.MergeOptions](int64(optsHandle))\n")
	b.WriteString("\t\tif !ok {\n\t\t\treturn C.CString(\"invalid MergeOptions handle\")\n\t\t}\n")
	b.WriteString("\t\topts = *optsPtr\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn toCError(repo.Merge(*ref, opts))\n}\n\n")
}

func generateWorktreeGo(pkg *Package, outputDir string) error {
	var wtType *HandleType
	for i := range pkg.Types {
		if pkg.Types[i].GoName == "Worktree" {
			wtType = &pkg.Types[i]
			break
		}
	}
	if wtType == nil {
		return fmt.Errorf("worktree type not found")
	}

	var b strings.Builder
	b.WriteString(`package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

`)

	for _, m := range wtType.Methods {
		generateWorktreeMethod(&b, m)
	}

	b.WriteString(`//export GitWorktreeFree
func GitWorktreeFree(wtHandle C.longlong) {
	removeHandle(int64(wtHandle))
}

`)

	// Suppress unused imports
	b.WriteString("var (\n")
	b.WriteString("\t_ = json.Marshal\n")
	b.WriteString("\t_ = time.Now\n")
	b.WriteString("\t_ object.Signature\n")
	b.WriteString("\t_ plumbing.Hash\n")
	b.WriteString(")\n")

	return writeGenFile(outputDir, "worktree_gen.go", b.String())
}

func wtLoadPreamble(b *strings.Builder) {
	b.WriteString("\twt, ok := loadHandle[*git.Worktree](int64(wtHandle))\n")
	b.WriteString("\tif !ok {\n\t\treturn C.CString(\"invalid worktree handle\")\n\t}\n")
}

func generateWorktreeMethod(b *strings.Builder, m Method) {
	switch m.GoName {
	case "Status":
		generateWtStatus(b, m)
	case "Add":
		generateWtAddPath(b, m)
	case "AddWithOptions":
		generateWtAddWithOptions(b, m)
	case "AddGlob":
		generateWtAddGlob(b, m)
	case "Commit":
		generateWtCommit(b, m)
	case "Checkout":
		generateWtCheckout(b, m)
	case "Pull":
		generateWtPull(b, m)
	case "Reset":
		generateWtReset(b, m)
	case "Restore":
		generateWtRestore(b, m)
	case "Clean":
		generateWtClean(b, m)
	case "Move":
		generateWtMove(b, m)
	case "Submodule":
		generateWtSubmodule(b, m)
	case "Submodules":
		generateWtSubmodules(b, m)
	}
}

func generateWtStatus(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, jsonOut **C.char) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString(`	status, err := wt.Status()
	if err != nil {
		return toCError(err)
	}
	type fileStatusJSON struct {
		Staging  string ` + "`json:\"staging\"`" + `
		Worktree string ` + "`json:\"worktree\"`" + `
		Extra    string ` + "`json:\"extra,omitempty\"`" + `
	}
	out := make(map[string]fileStatusJSON, len(status))
	for path, fs := range status {
		out[path] = fileStatusJSON{
			Staging:  string(fs.Staging),
			Worktree: string(fs.Worktree),
			Extra:    fs.Extra,
		}
	}
	data, err := json.Marshal(out)
	if err != nil {
		return toCError(err)
	}
	*jsonOut = C.CString(string(data))
	return nil
}

`)
}

func generateWtAddPath(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, path *C.char, hashOut **C.char) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\thash, err := wt.Add(C.GoString(path))\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*hashOut = C.CString(hash.String())\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateWtAddWithOptions(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, optsHandle C.longlong) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\topts, ok2 := loadHandle[*git.AddOptions](int64(optsHandle))\n")
	b.WriteString("\tif !ok2 {\n\t\treturn C.CString(\"invalid AddOptions handle\")\n\t}\n")
	b.WriteString("\treturn toCError(wt.AddWithOptions(opts))\n}\n\n")
}

func generateWtAddGlob(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, pattern *C.char) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\treturn toCError(wt.AddGlob(C.GoString(pattern)))\n}\n\n")
}

func generateWtCommit(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, msg *C.char, optsHandle C.longlong, hashOut **C.char) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\tvar opts *git.CommitOptions\n")
	b.WriteString("\tif int64(optsHandle) != 0 {\n")
	b.WriteString("\t\tvar ok2 bool\n")
	b.WriteString("\t\topts, ok2 = loadHandle[*git.CommitOptions](int64(optsHandle))\n")
	b.WriteString("\t\tif !ok2 {\n\t\t\treturn C.CString(\"invalid CommitOptions handle\")\n\t\t}\n")
	b.WriteString("\t} else {\n")
	b.WriteString("\t\topts = &git.CommitOptions{}\n")
	b.WriteString("\t}\n")
	b.WriteString("\thash, err := wt.Commit(C.GoString(msg), opts)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*hashOut = C.CString(hash.String())\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateWtCheckout(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, optsHandle C.longlong) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\topts, ok2 := loadHandle[*git.CheckoutOptions](int64(optsHandle))\n")
	b.WriteString("\tif !ok2 {\n\t\treturn C.CString(\"invalid CheckoutOptions handle\")\n\t}\n")
	b.WriteString("\treturn toCError(wt.Checkout(opts))\n}\n\n")
}

func generateWtPull(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, optsHandle C.longlong) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\topts, ok2 := loadHandle[*git.PullOptions](int64(optsHandle))\n")
	b.WriteString("\tif !ok2 {\n\t\treturn C.CString(\"invalid PullOptions handle\")\n\t}\n")
	b.WriteString("\treturn toCError(wt.Pull(opts))\n}\n\n")
}

func generateWtReset(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, optsHandle C.longlong) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\topts, ok2 := loadHandle[*git.ResetOptions](int64(optsHandle))\n")
	b.WriteString("\tif !ok2 {\n\t\treturn C.CString(\"invalid ResetOptions handle\")\n\t}\n")
	b.WriteString("\treturn toCError(wt.Reset(opts))\n}\n\n")
}

func generateWtRestore(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, optsHandle C.longlong) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\topts, ok2 := loadHandle[*git.RestoreOptions](int64(optsHandle))\n")
	b.WriteString("\tif !ok2 {\n\t\treturn C.CString(\"invalid RestoreOptions handle\")\n\t}\n")
	b.WriteString("\treturn toCError(wt.Restore(opts))\n}\n\n")
}

func generateWtClean(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, optsHandle C.longlong) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\topts, ok2 := loadHandle[*git.CleanOptions](int64(optsHandle))\n")
	b.WriteString("\tif !ok2 {\n\t\treturn C.CString(\"invalid CleanOptions handle\")\n\t}\n")
	b.WriteString("\treturn toCError(wt.Clean(opts))\n}\n\n")
}

func generateWtMove(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, fromPath *C.char, toPath *C.char, hashOut **C.char) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\thash, err := wt.Move(C.GoString(fromPath), C.GoString(toPath))\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*hashOut = C.CString(hash.String())\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateWtSubmodule(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, name *C.char, handleOut *C.longlong) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\tsub, err := wt.Submodule(C.GoString(name))\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*handleOut = C.longlong(storeHandle(sub))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateWtSubmodules(b *strings.Builder, m Method) {
	fmt.Fprintf(b, "//export %s\n", m.CName)
	fmt.Fprintf(b, "func %s(wtHandle C.longlong, jsonOut **C.char) *C.char {\n", m.CName)
	wtLoadPreamble(b)
	b.WriteString("\tsubs, err := wt.Submodules()\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\tnames := make([]string, len(subs))\n")
	b.WriteString("\tfor i, s := range subs {\n")
	b.WriteString("\t\tnames[i] = s.Config().Name\n")
	b.WriteString("\t}\n")
	b.WriteString("\tdata, err := json.Marshal(names)\n")
	b.WriteString("\tif err != nil {\n\t\treturn toCError(err)\n\t}\n")
	b.WriteString("\t*jsonOut = C.CString(string(data))\n")
	b.WriteString("\treturn nil\n}\n\n")
}

func generateRemoteGo(pkg *Package, outputDir string) error {
	content := `package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"

	git "github.com/go-git/go-git/v6"
)

//export GitRemoteFetch
func GitRemoteFetch(remoteHandle C.longlong, optsHandle C.longlong) *C.char {
	remote, ok := loadHandle[*git.Remote](int64(remoteHandle))
	if !ok {
		return C.CString("invalid remote handle")
	}
	opts, ok := loadHandle[*git.FetchOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid FetchOptions handle")
	}
	return toCError(remote.Fetch(opts))
}

//export GitRemotePush
func GitRemotePush(remoteHandle C.longlong, optsHandle C.longlong) *C.char {
	remote, ok := loadHandle[*git.Remote](int64(remoteHandle))
	if !ok {
		return C.CString("invalid remote handle")
	}
	opts, ok := loadHandle[*git.PushOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid PushOptions handle")
	}
	return toCError(remote.Push(opts))
}

//export GitRemoteList
func GitRemoteList(remoteHandle C.longlong, optsHandle C.longlong, jsonOut **C.char) *C.char {
	remote, ok := loadHandle[*git.Remote](int64(remoteHandle))
	if !ok {
		return C.CString("invalid remote handle")
	}
	opts, ok := loadHandle[*git.ListOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid ListOptions handle")
	}
	refs, err := remote.List(opts)
	if err != nil {
		return toCError(err)
	}
	type refJSON struct {
		Name string ` + "`json:\"name\"`" + `
		Hash string ` + "`json:\"hash\"`" + `
	}
	out := make([]refJSON, len(refs))
	for i, r := range refs {
		out[i] = refJSON{Name: string(r.Name()), Hash: r.Hash().String()}
	}
	data, err := json.Marshal(out)
	if err != nil {
		return toCError(err)
	}
	*jsonOut = C.CString(string(data))
	return nil
}

//export GitRemoteConfigName
func GitRemoteConfigName(remoteHandle C.longlong, nameOut **C.char) *C.char {
	remote, ok := loadHandle[*git.Remote](int64(remoteHandle))
	if !ok {
		return C.CString("invalid remote handle")
	}
	*nameOut = C.CString(remote.Config().Name)
	return nil
}

//export GitRemoteFree
func GitRemoteFree(remoteHandle C.longlong) {
	removeHandle(int64(remoteHandle))
}
`
	return writeGenFile(outputDir, "remote_gen.go", content)
}

func generateSubmoduleGo(pkg *Package, outputDir string) error {
	content := `package main

/*
#include <stdlib.h>
*/
import "C"
import (
	git "github.com/go-git/go-git/v6"
)

//export GitSubmoduleInit
func GitSubmoduleInit(subHandle C.longlong) *C.char {
	sub, ok := loadHandle[*git.Submodule](int64(subHandle))
	if !ok {
		return C.CString("invalid submodule handle")
	}
	return toCError(sub.Init())
}

//export GitSubmoduleUpdate
func GitSubmoduleUpdate(subHandle C.longlong, optsHandle C.longlong) *C.char {
	sub, ok := loadHandle[*git.Submodule](int64(subHandle))
	if !ok {
		return C.CString("invalid submodule handle")
	}
	opts, ok := loadHandle[*git.SubmoduleUpdateOptions](int64(optsHandle))
	if !ok {
		return C.CString("invalid SubmoduleUpdateOptions handle")
	}
	return toCError(sub.Update(opts))
}

//export GitSubmoduleRepository
func GitSubmoduleRepository(subHandle C.longlong, handleOut *C.longlong) *C.char {
	sub, ok := loadHandle[*git.Submodule](int64(subHandle))
	if !ok {
		return C.CString("invalid submodule handle")
	}
	repo, err := sub.Repository()
	if err != nil {
		return toCError(err)
	}
	*handleOut = C.longlong(storeHandle(repo))
	return nil
}

//export GitSubmoduleConfigName
func GitSubmoduleConfigName(subHandle C.longlong, nameOut **C.char) *C.char {
	sub, ok := loadHandle[*git.Submodule](int64(subHandle))
	if !ok {
		return C.CString("invalid submodule handle")
	}
	*nameOut = C.CString(sub.Config().Name)
	return nil
}

//export GitSubmoduleFree
func GitSubmoduleFree(subHandle C.longlong) {
	removeHandle(int64(subHandle))
}
`
	return writeGenFile(outputDir, "submodule_gen.go", content)
}

func generateIteratorsGo(outputDir string) error {
	content := `package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"io"

	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/storer"
)

//export GitCommitIterNext
func GitCommitIterNext(iterHandle C.longlong, hashOut **C.char, msgOut **C.char, authorNameOut **C.char, authorEmailOut **C.char, tsOut *C.longlong, eofOut *C.int) *C.char {
	iter, ok := loadHandle[object.CommitIter](int64(iterHandle))
	if !ok {
		return C.CString("invalid iterator handle")
	}
	commit, err := iter.Next()
	if err == io.EOF || errors.Is(err, plumbing.ErrObjectNotFound) {
		*eofOut = 1
		return nil
	}
	if err != nil {
		return toCError(err)
	}
	*eofOut = 0
	*hashOut = C.CString(commit.Hash.String())
	*msgOut = C.CString(commit.Message)
	*authorNameOut = C.CString(commit.Author.Name)
	*authorEmailOut = C.CString(commit.Author.Email)
	*tsOut = C.longlong(commit.Author.When.Unix())
	return nil
}

//export GitCommitIterFree
func GitCommitIterFree(iterHandle C.longlong) {
	iter, ok := loadHandle[object.CommitIter](int64(iterHandle))
	if ok {
		iter.Close()
	}
	removeHandle(int64(iterHandle))
}

//export GitReferenceIterNext
func GitReferenceIterNext(iterHandle C.longlong, refNameOut **C.char, hashOut **C.char, eofOut *C.int) *C.char {
	iter, ok := loadHandle[storer.ReferenceIter](int64(iterHandle))
	if !ok {
		return C.CString("invalid iterator handle")
	}
	ref, err := iter.Next()
	if err == io.EOF || errors.Is(err, plumbing.ErrObjectNotFound) {
		*eofOut = 1
		return nil
	}
	if err != nil {
		return toCError(err)
	}
	*eofOut = 0
	*refNameOut = C.CString(string(ref.Name()))
	*hashOut = C.CString(ref.Hash().String())
	return nil
}

//export GitReferenceIterFree
func GitReferenceIterFree(iterHandle C.longlong) {
	iter, ok := loadHandle[storer.ReferenceIter](int64(iterHandle))
	if ok {
		iter.Close()
	}
	removeHandle(int64(iterHandle))
}

var _ plumbing.Hash
`
	return writeGenFile(outputDir, "iterators_gen.go", content)
}
