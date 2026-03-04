package main

// Package represents a parsed Go package's public API surface.
type Package struct {
	Name      string
	Functions []Function
	Types     []HandleType
	Options   []OptionsStruct
}

// HandleType represents a Go type that maps to an opaque int64 handle in C.
type HandleType struct {
	GoName    string
	CPrefix   string
	Methods   []Method
	IsPointer bool
}

// Method represents an exported method on a handle type.
type Method struct {
	GoName     string
	CName      string
	Receiver   string
	Params     []Param
	Returns    []Return
	HasContext bool
	DocComment string
}

// Function represents a top-level exported function.
type Function struct {
	GoName     string
	CName      string
	Params     []Param
	Returns    []Return
	DocComment string
}

// Param represents a function/method parameter.
type Param struct {
	GoName     string
	GoType     string
	CName      string
	CType      string
	CSharpType string
	Mapping    TypeMapping
}

// Return represents a function/method return value.
type Return struct {
	GoType     string
	CType      string
	CSharpType string
	Mapping    TypeMapping
	IsError    bool
}

// OptionsStruct represents a Go options struct that maps to handle+setter pattern.
type OptionsStruct struct {
	GoName  string
	CPrefix string
	Fields  []OptionsField
}

// OptionsField represents a single field in an options struct.
type OptionsField struct {
	GoName      string
	GoType      string
	CSetterName string
	CType       string
	CSharpType  string
	Mapping     TypeMapping
}

// TypeMapping describes how a Go type maps to C and C# types.
type TypeMapping struct {
	Kind       MappingKind
	GoType     string
	CType      string
	CSharpType string
	HandleType string
	GoToC      string
	CToGo      string
}

// MappingKind classifies how a type is marshalled across the FFI boundary.
type MappingKind int

const (
	MappingPrimitive MappingKind = iota
	MappingString
	MappingBool
	MappingHash
	MappingHandle
	MappingOptions
	MappingAuth
	MappingSigner
	MappingIterator
	MappingCallback
	MappingByteSlice
	MappingEnum
	MappingReferenceName
	MappingSkip
)
