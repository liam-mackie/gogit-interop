package main

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
