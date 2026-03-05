package main

import (
	"fmt"
	"strings"
)

func generateAuthGo(outputDir string) error {
	content := genHeader + `
/*
#include <stdlib.h>
#include "callbacks.h"
*/
import "C"
import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"unsafe"

	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	stdssh "golang.org/x/crypto/ssh"
)

func setSSHHostKeyCallback(authHandle C.longlong, cb stdssh.HostKeyCallback) *C.char {
	h, ok := loadHandle[interface{}](int64(authHandle))
	if !ok {
		return C.CString("invalid auth handle")
	}
	switch a := h.(type) {
	case *ssh.PublicKeys:
		a.HostKeyCallback = cb
	case *ssh.PublicKeysCallback:
		a.HostKeyCallback = cb
	case *ssh.Password:
		a.HostKeyCallback = cb
	case *ssh.PasswordCallback:
		a.HostKeyCallback = cb
	case *ssh.KeyboardInteractive:
		a.HostKeyCallback = cb
	default:
		return C.CString(fmt.Sprintf("auth type %T does not support HostKeyCallback", h))
	}
	return nil
}

//export GitAuthSetInsecureIgnoreHostKey
func GitAuthSetInsecureIgnoreHostKey(authHandle C.longlong) *C.char {
	return setSSHHostKeyCallback(authHandle, stdssh.InsecureIgnoreHostKey())
}

//export GitAuthSetKnownHostsFiles
func GitAuthSetKnownHostsFiles(authHandle C.longlong, filesJSON *C.char) *C.char {
	var paths []string
	if err := json.Unmarshal([]byte(C.GoString(filesJSON)), &paths); err != nil {
		return toCError(err)
	}
	cb, err := ssh.NewKnownHostsCallback(paths...)
	if err != nil {
		return toCError(err)
	}
	return setSSHHostKeyCallback(authHandle, cb)
}

//export GitAuthSetHostKeyCallback
func GitAuthSetHostKeyCallback(authHandle C.longlong, fn C.GitHostKeyFunc, userData unsafe.Pointer) *C.char {
	cb := func(hostname string, remote net.Addr, key stdssh.PublicKey) error {
		keyType := key.Type()
		keyBytes := base64.StdEncoding.EncodeToString(key.Marshal())

		cHostname := C.CString(hostname)
		defer C.free(unsafe.Pointer(cHostname))
		cRemote := C.CString(remote.String())
		defer C.free(unsafe.Pointer(cRemote))
		cKeyType := C.CString(keyType)
		defer C.free(unsafe.Pointer(cKeyType))
		cKeyB64 := C.CString(keyBytes)
		defer C.free(unsafe.Pointer(cKeyB64))

		errStr := C.callHostKeyFunc(fn, cHostname, cRemote, cKeyType, cKeyB64, userData)
		if errStr != nil {
			defer C.free(unsafe.Pointer(errStr))
			return errors.New(C.GoString(errStr))
		}
		return nil
	}
	return setSSHHostKeyCallback(authHandle, cb)
}

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
	content := genHeader + `
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

func generateIteratorsGo(outputDir string) error {
	content := genHeader + `
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

func generateSignatureSetters(b *strings.Builder, cPrefix, goType, fieldName string) {
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
