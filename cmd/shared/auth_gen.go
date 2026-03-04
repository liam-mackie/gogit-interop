package main

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
