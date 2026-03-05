package main

import "go/types"

var excludedTypes = map[string]bool{
	"github.com/go-git/go-billy/v6.Filesystem":                true,
	"github.com/go-git/go-git/v6/storage.Storer":              true,
	"github.com/go-git/go-git/v6/plumbing/format/packfile":    true,
	"github.com/go-git/go-git/v6/plumbing/protocol/packp":     true,
	"github.com/go-git/go-git/v6/plumbing/transport.Endpoint": true,
}

var excludedFieldTypes = map[string]bool{
	"github.com/go-git/go-billy/v6.Filesystem":                              true,
	"github.com/go-git/go-git/v6/storage.Storer":                            true,
	"github.com/go-git/go-git/v6/plumbing/protocol/packp.Filter":            true,
	"github.com/go-git/go-git/v6/plumbing/protocol/packp/sideband.Progress": true,
	"io.Writer": true,
}

func isExcludedType(qname string) bool {
	return excludedTypes[qname]
}

func isExcludedFieldType(qname string) bool {
	return excludedFieldTypes[qname]
}

func containsChannel(t types.Type) bool {
	switch u := t.Underlying().(type) {
	case *types.Chan:
		return true
	case *types.Struct:
		for i := range u.NumFields() {
			if containsChannel(u.Field(i).Type()) {
				return true
			}
		}
	}
	return false
}

func isFuncType(t types.Type) bool {
	_, ok := t.Underlying().(*types.Signature)
	return ok
}
