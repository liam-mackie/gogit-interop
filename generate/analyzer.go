package main

import (
	"fmt"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

var packagesToLoad = []string{
	"github.com/go-git/go-git/v6",
	"github.com/go-git/go-git/v6/plumbing",
	"github.com/go-git/go-git/v6/plumbing/object",
	"github.com/go-git/go-git/v6/plumbing/storer",
	"github.com/go-git/go-git/v6/plumbing/transport",
	"github.com/go-git/go-git/v6/config",
}

type analyzer struct {
	pkgMap            map[string]*packages.Package
	discoveredHandles map[string]*HandleType
	discoveredOptions map[string]*OptionsStruct
	discoveredEnums   map[string]*EnumType
	handleQueue       []namedType
}

type namedType struct {
	name string
	obj  *types.Named
}

func analyzePackages() *Package {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
	}

	pkgs, err := packages.Load(cfg, packagesToLoad...)
	if err != nil {
		panic(fmt.Sprintf("failed to load packages: %v", err))
	}

	a := &analyzer{
		pkgMap:            make(map[string]*packages.Package),
		discoveredHandles: make(map[string]*HandleType),
		discoveredOptions: make(map[string]*OptionsStruct),
		discoveredEnums:   make(map[string]*EnumType),
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				fmt.Printf("WARNING: package %s: %v\n", pkg.PkgPath, e)
			}
		}
		a.pkgMap[pkg.PkgPath] = pkg
	}

	a.discoverEnums()
	a.discoverOptions()
	a.registerSeedHandleTypes()
	functions := a.discoverFunctions()
	a.processHandleQueue()

	result := &Package{Name: "gogit"}
	result.Functions = functions

	handleNames := sortedKeys(a.discoveredHandles)
	for _, name := range handleNames {
		result.Types = append(result.Types, *a.discoveredHandles[name])
	}

	optNames := sortedKeys(a.discoveredOptions)
	for _, name := range optNames {
		result.Options = append(result.Options, *a.discoveredOptions[name])
	}

	enumNames := sortedKeys(a.discoveredEnums)
	for _, name := range enumNames {
		result.Enums = append(result.Enums, *a.discoveredEnums[name])
	}

	return result
}

func (a *analyzer) discoverEnums() {
	for _, pkg := range a.pkgMap {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			tn, ok := obj.(*types.TypeName)
			if !ok || !tn.Exported() {
				continue
			}
			named, ok := tn.Type().(*types.Named)
			if !ok {
				continue
			}
			basic, ok := named.Underlying().(*types.Basic)
			if !ok {
				continue
			}
			if basic.Info()&types.IsInteger == 0 {
				continue
			}
			qname := pkg.PkgPath + "." + name
			if a.hasConstants(pkg, named) {
				cPrefix := "Git" + name
				a.discoveredEnums[qname] = &EnumType{
					GoName:     name,
					ImportPath: pkg.PkgPath,
					CPrefix:    cPrefix,
					Values:     a.collectConstants(pkg, named),
				}
				registerEnumMapping(qname)
			}
		}
	}
}

func (a *analyzer) hasConstants(pkg *packages.Package, named *types.Named) bool {
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		c, ok := obj.(*types.Const)
		if !ok {
			continue
		}
		if types.Identical(c.Type(), named) {
			return true
		}
	}
	return false
}

func (a *analyzer) collectConstants(pkg *packages.Package, named *types.Named) []EnumValue {
	var values []EnumValue
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		c, ok := obj.(*types.Const)
		if !ok || !c.Exported() {
			continue
		}
		if !types.Identical(c.Type(), named) {
			continue
		}
		val, _ := constantToInt64(c)
		values = append(values, EnumValue{
			GoName: name,
			CName:  "Git" + name,
			Value:  val,
		})
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Value < values[j].Value })
	return values
}

func (a *analyzer) discoverOptions() {
	for _, pkg := range a.pkgMap {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			tn, ok := obj.(*types.TypeName)
			if !ok || !tn.Exported() {
				continue
			}
			if !strings.HasSuffix(name, "Options") {
				continue
			}
			named, ok := tn.Type().(*types.Named)
			if !ok {
				continue
			}
			st, ok := named.Underlying().(*types.Struct)
			if !ok {
				continue
			}

			qname := pkg.PkgPath + "." + name
			if isExcludedType(qname) {
				continue
			}

			opts := &OptionsStruct{
				GoName:     name,
				CPrefix:    "Git" + name,
				ImportPath: pkg.PkgPath,
			}

			for i := range st.NumFields() {
				field := st.Field(i)
				if !field.Exported() {
					continue
				}

				fieldQName := qualifiedTypeName(field.Type())
				if isExcludedFieldType(fieldQName) {
					continue
				}
				if isFuncType(field.Type()) {
					continue
				}
				if containsChannel(field.Type()) {
					continue
				}

				m := resolveTypeMapping(field.Type())
				if m.Kind == MappingSkip {
					fmt.Printf("WARNING: skipping %s.%s — unmappable type %s\n", name, field.Name(), field.Type())
					continue
				}

				opts.Fields = append(opts.Fields, OptionsField{
					GoName:      field.Name(),
					GoType:      m.GoType,
					CSetterName: lowerFirst(field.Name()),
					CType:       m.CType,
					CSharpType:  m.CSharpType,
					Mapping:     m,
				})
			}

			ptrQName := pkg.PkgPath + ".*" + name
			registerOptionsMapping(ptrQName, name)
			a.discoveredOptions[qname] = opts
		}
	}
}

func (a *analyzer) discoverFunctions() []Function {
	gitPkg, ok := a.pkgMap["github.com/go-git/go-git/v6"]
	if !ok {
		return nil
	}

	var functions []Function
	scope := gitPkg.Types.Scope()

	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		fn, ok := obj.(*types.Func)
		if !ok || !fn.Exported() {
			continue
		}
		sig := fn.Type().(*types.Signature)
		if sig.Recv() != nil {
			continue
		}

		params, paramsOk := a.mapTupleToParams(sig.Params())
		if !paramsOk && sig.Variadic() {
			params, paramsOk = a.mapTupleToParamsSkipVariadic(sig.Params())
		}
		returns, returnsOk := a.mapTupleToReturns(sig.Results())
		if !paramsOk || !returnsOk {
			fmt.Printf("WARNING: skipping function %s — unmappable signature\n", name)
			continue
		}

		cName := "Git" + name
		functions = append(functions, Function{
			GoName:     name,
			CName:      cName,
			Params:     params,
			Returns:    returns,
			HasContext: tupleHasContext(sig.Params()),
		})
	}

	sort.Slice(functions, func(i, j int) bool { return functions[i].GoName < functions[j].GoName })
	return functions
}

func (a *analyzer) registerSeedHandleTypes() {
	seedTypes := []struct {
		pkgPath string
		name    string
	}{
		{"github.com/go-git/go-git/v6", "Repository"},
		{"github.com/go-git/go-git/v6", "Worktree"},
		{"github.com/go-git/go-git/v6", "Remote"},
		{"github.com/go-git/go-git/v6", "Submodule"},
		{"github.com/go-git/go-git/v6/plumbing/object", "Commit"},
		{"github.com/go-git/go-git/v6/plumbing/object", "Tree"},
		{"github.com/go-git/go-git/v6/plumbing/object", "Blob"},
		{"github.com/go-git/go-git/v6/plumbing/object", "Tag"},
		{"github.com/go-git/go-git/v6/plumbing/object", "File"},
	}

	for _, seed := range seedTypes {
		pkg, ok := a.pkgMap[seed.pkgPath]
		if !ok {
			continue
		}
		obj := pkg.Types.Scope().Lookup(seed.name)
		if obj == nil {
			continue
		}
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}

		qname := seed.pkgPath + ".*" + seed.name
		if _, exists := a.discoveredHandles[qname]; exists {
			continue
		}

		a.enqueueHandle(qname, named)
	}
}

func (a *analyzer) enqueueHandle(qname string, named *types.Named) {
	if _, exists := a.discoveredHandles[qname]; exists {
		return
	}

	typeName := named.Obj().Name()
	ht := &HandleType{
		GoName:     typeName,
		CPrefix:    "Git" + typeName,
		ImportPath: named.Obj().Pkg().Path(),
		IsPointer:  true,
	}
	a.discoveredHandles[qname] = ht
	registerHandleMapping(qname, typeName)

	a.handleQueue = append(a.handleQueue, namedType{name: qname, obj: named})
}

func (a *analyzer) processHandleQueue() {
	for len(a.handleQueue) > 0 {
		item := a.handleQueue[0]
		a.handleQueue = a.handleQueue[1:]
		a.processHandleType(item.name, item.obj)
	}
}

func (a *analyzer) processHandleType(qname string, named *types.Named) {
	ht := a.discoveredHandles[qname]

	for i := range named.NumMethods() {
		m := named.Method(i)
		if !m.Exported() {
			continue
		}
		sig := m.Type().(*types.Signature)

		params, paramsOk := a.mapTupleToParams(sig.Params())
		if !paramsOk && sig.Variadic() {
			params, paramsOk = a.mapTupleToParamsSkipVariadic(sig.Params())
		}
		returns, returnsOk := a.mapTupleToReturns(sig.Results())

		override := isOverrideMethod(ht.GoName, m.Name())
		if !paramsOk || !returnsOk {
			if !override {
				fmt.Printf("WARNING: skipping %s.%s — unmappable signature\n", ht.GoName, m.Name())
				continue
			}
			params = nil
			returns = nil
		}

		hasCtx := tupleHasContext(sig.Params())
		markOptionalParams(params)

		cName := ht.CPrefix + m.Name()
		method := Method{
			GoName:     m.Name(),
			CName:      cName,
			Receiver:   ht.GoName,
			Params:     params,
			Returns:    returns,
			HasContext: hasCtx,
		}
		ht.Methods = append(ht.Methods, method)
	}

	sort.Slice(ht.Methods, func(i, j int) bool { return ht.Methods[i].GoName < ht.Methods[j].GoName })

	st, ok := named.Underlying().(*types.Struct)
	if ok {
		for i := range st.NumFields() {
			field := st.Field(i)
			if !field.Exported() {
				continue
			}
			m := resolveTypeMapping(field.Type())
			if m.Kind == MappingSkip {
				continue
			}
			ht.Fields = append(ht.Fields, HandleField{
				GoName:      field.Name(),
				GoType:      field.Type().String(),
				CGetterName: ht.CPrefix + "Get" + field.Name(),
				Mapping:     m,
			})
		}
	}
}

func (a *analyzer) mapTupleToParamsSkipVariadic(tuple *types.Tuple) ([]Param, bool) {
	if tuple.Len() == 0 {
		return nil, true
	}
	nonVariadic := tuple.Len() - 1
	var params []Param
	for i := range nonVariadic {
		v := tuple.At(i)
		t := v.Type()

		if isContextType(t) {
			continue
		}

		m := resolveTypeMapping(t)
		if m.Kind == MappingSkip {
			return nil, false
		}

		name := v.Name()
		if name == "" {
			name = fmt.Sprintf("p%d", i)
		}

		cName := name
		if m.Kind == MappingOptions || m.Kind == MappingHandle {
			cName = name + "Handle"
		}

		params = append(params, Param{
			GoName:     name,
			GoType:     m.GoType,
			CName:      cName,
			CType:      m.CType,
			CSharpType: m.CSharpType,
			Mapping:    m,
		})
	}
	return params, true
}

func (a *analyzer) mapTupleToParams(tuple *types.Tuple) ([]Param, bool) {
	var params []Param
	for i := range tuple.Len() {
		v := tuple.At(i)
		t := v.Type()

		if isContextType(t) {
			continue
		}

		m := resolveTypeMapping(t)
		if m.Kind == MappingSkip {
			return nil, false
		}

		name := v.Name()
		if name == "" {
			name = fmt.Sprintf("p%d", i)
		}

		cName := name
		if m.Kind == MappingOptions || m.Kind == MappingHandle {
			cName = name + "Handle"
		}

		params = append(params, Param{
			GoName:     name,
			GoType:     m.GoType,
			CName:      cName,
			CType:      m.CType,
			CSharpType: m.CSharpType,
			Mapping:    m,
		})
	}
	return params, true
}

func (a *analyzer) mapTupleToReturns(tuple *types.Tuple) ([]Return, bool) {
	var returns []Return
	for i := range tuple.Len() {
		v := tuple.At(i)
		t := v.Type()

		if isErrorType(t) {
			returns = append(returns, Return{
				GoType:  "error",
				IsError: true,
				Mapping: resolveTypeMapping(t),
			})
			continue
		}

		m := resolveTypeMapping(t)
		if m.Kind == MappingSkip {
			return nil, false
		}

		returns = append(returns, Return{
			GoType:     m.GoType,
			CType:      m.CType,
			CSharpType: m.CSharpType,
			Mapping:    m,
		})
	}
	return returns, true
}

func markOptionalParams(params []Param) {
	if len(params) == 0 {
		return
	}
	last := &params[len(params)-1]
	if last.Mapping.Kind == MappingOptions {
		last.IsOptional = true
	}
}

func tupleHasContext(tuple *types.Tuple) bool {
	for i := range tuple.Len() {
		if isContextType(tuple.At(i).Type()) {
			return true
		}
	}
	return false
}

func isContextType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() != nil && obj.Pkg().Path() == "context" && obj.Name() == "Context"
}

func isErrorType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if ok {
		return named.Obj().Pkg() == nil && named.Obj().Name() == "error"
	}
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return iface.NumMethods() == 1 && iface.Method(0).Name() == "Error"
}

func qualifiedTypeName(t types.Type) string {
	switch v := t.(type) {
	case *types.Named:
		if v.Obj().Pkg() != nil {
			return v.Obj().Pkg().Path() + "." + v.Obj().Name()
		}
		return v.Obj().Name()
	case *types.Pointer:
		return qualifiedTypeName(v.Elem())
	default:
		return t.String()
	}
}

func constantToInt64(c *types.Const) (int64, bool) {
	if c.Val().Kind() == 2 { // constant.Int
		val, ok := c.Val().ExactString(), true
		if !ok {
			return 0, false
		}
		var n int64
		fmt.Sscanf(val, "%d", &n)
		return n, true
	}
	return 0, false
}

func lowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func sortedKeys[V any](m map[string]*V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
