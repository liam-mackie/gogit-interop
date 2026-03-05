package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var optionsTemplate = template.Must(template.New("options").Parse(csGenHeader + `#nullable enable
namespace GoGit.Interop;

public sealed class {{.ClassName}} : IDisposable
{
    private long _handle;
    private bool _disposed;

    internal long Handle => _handle;

    public {{.ClassName}}()
    {
        NativeMethods.{{.CPrefix}}New(out _handle);
    }
{{range .Fields}}
    public {{.ClassName}} {{.MethodName}}({{.CSharpParamType}} value)
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.{{.SetterName}}(_handle, {{.MarshalExpr}}));
        return this;
    }
{{end}}{{range .ExtraSetters}}
    public {{.ClassName}} {{.MethodName}}({{.Params}})
    {
        ObjectDisposedException.ThrowIf(_disposed, this);
        NativeMethods.ThrowIfError(NativeMethods.{{.NativeMethod}}(_handle, {{.Args}}));
        return this;
    }
{{end}}
    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;
        NativeMethods.{{.CPrefix}}Free(_handle);
        _handle = 0;
    }
}
`))

type optionsTemplateData struct {
	ClassName    string
	CPrefix      string
	Fields       []optionsFieldData
	ExtraSetters []extraSetterData
}

type optionsFieldData struct {
	ClassName       string
	MethodName      string
	CSharpParamType string
	SetterName      string
	MarshalExpr     string
}

type extraSetterData struct {
	ClassName    string
	MethodName   string
	Params       string
	NativeMethod string
	Args         string
}

func generateCSharpOptions(pkg *Package, dir string) error {
	optionsDir := filepath.Join(dir, "Options")

	for _, opts := range pkg.Options {
		data := optionsTemplateData{
			ClassName: opts.GoName,
			CPrefix:   opts.CPrefix,
		}

		for _, f := range opts.Fields {
			fd := optionsFieldData{
				ClassName:  opts.GoName,
				MethodName: "Set" + f.GoName,
				SetterName: opts.CPrefix + "Set" + f.GoName,
			}

			switch f.Mapping.Kind {
			case MappingString, MappingReferenceName, MappingHash:
				fd.CSharpParamType = "string"
				fd.MarshalExpr = "value"
			case MappingBool:
				fd.CSharpParamType = "bool"
				fd.MarshalExpr = "value ? 1 : 0"
			case MappingPrimitive, MappingEnum:
				fd.CSharpParamType = f.Mapping.CSharpType
				fd.MarshalExpr = "value"
			case MappingAuth:
				fd.CSharpParamType = "Auth"
				fd.MarshalExpr = "value.Handle"
			case MappingSigner:
				fd.CSharpParamType = "Signer"
				fd.MarshalExpr = "value.Handle"
			case MappingTime:
				fd.CSharpParamType = "DateTimeOffset"
				fd.MarshalExpr = "value.ToUnixTimeSeconds()"
			case MappingDuration:
				fd.CSharpParamType = "TimeSpan"
				fd.MarshalExpr = "value.Ticks * 100"
			case MappingStringSlice:
				fd.CSharpParamType = "string[]"
				fd.MarshalExpr = "System.Text.Json.JsonSerializer.Serialize(value)"
			default:
				continue
			}

			data.Fields = append(data.Fields, fd)
		}

		if opts.GoName == "CommitOptions" {
			data.ExtraSetters = append(data.ExtraSetters,
				extraSetterData{ClassName: opts.GoName, MethodName: "SetAuthor", Params: "string name, string email",
					NativeMethod: opts.CPrefix + "SetAuthorNameEmail", Args: "name, email"},
				extraSetterData{ClassName: opts.GoName, MethodName: "SetCommitter", Params: "string name, string email",
					NativeMethod: opts.CPrefix + "SetCommitterNameEmail", Args: "name, email"},
			)
		}
		if opts.GoName == "CreateTagOptions" {
			data.ExtraSetters = append(data.ExtraSetters,
				extraSetterData{ClassName: opts.GoName, MethodName: "SetTagger", Params: "string name, string email",
					NativeMethod: opts.CPrefix + "SetTaggerNameEmail", Args: "name, email"},
			)
		}
		if opts.GoName == "RestoreOptions" {
			data.ExtraSetters = append(data.ExtraSetters,
				extraSetterData{ClassName: opts.GoName, MethodName: "AddFile", Params: "string path",
					NativeMethod: opts.CPrefix + "AddFile", Args: "path"},
			)
		}

		var b strings.Builder
		if err := optionsTemplate.Execute(&b, data); err != nil {
			return fmt.Errorf("executing template for %s: %w", opts.GoName, err)
		}

		if err := os.WriteFile(filepath.Join(optionsDir, opts.GoName+".cs"), []byte(b.String()), 0644); err != nil {
			return err
		}
	}

	return nil
}

func nativeMethodsOptionsSection(b *strings.Builder, pkg *Package) {
	for _, opts := range pkg.Options {
		fmt.Fprintf(b, "    // %s\n\n", opts.GoName)
		fmt.Fprintf(b, "    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
		fmt.Fprintf(b, "    public static extern void %sNew(out long handleOut);\n\n", opts.CPrefix)

		for _, f := range opts.Fields {
			setterName := opts.CPrefix + "Set" + f.GoName

			switch f.Mapping.Kind {
			case MappingString, MappingReferenceName, MappingHash, MappingStringSlice:
				fmt.Fprintf(b, "    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
				fmt.Fprintf(b, "    public static extern IntPtr %s(long handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string val);\n\n", setterName)
			case MappingBool:
				fmt.Fprintf(b, "    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
				fmt.Fprintf(b, "    public static extern IntPtr %s(long handle, int val);\n\n", setterName)
			case MappingPrimitive, MappingEnum:
				csharpType := "int"
				if f.Mapping.CSharpType == "long" {
					csharpType = "long"
				} else if f.Mapping.CSharpType == "uint" {
					csharpType = "uint"
				}
				fmt.Fprintf(b, "    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
				fmt.Fprintf(b, "    public static extern IntPtr %s(long handle, %s val);\n\n", setterName, csharpType)
			case MappingAuth:
				fmt.Fprintf(b, "    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
				fmt.Fprintf(b, "    public static extern IntPtr %s(long handle, long authHandle);\n\n", setterName)
			case MappingSigner:
				fmt.Fprintf(b, "    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
				fmt.Fprintf(b, "    public static extern IntPtr %s(long handle, long signerHandle);\n\n", setterName)
			case MappingTime, MappingDuration:
				fmt.Fprintf(b, "    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
				fmt.Fprintf(b, "    public static extern IntPtr %s(long handle, long val);\n\n", setterName)
			}
		}

		if opts.GoName == "CommitOptions" {
			for _, sigField := range []string{"Author", "Committer"} {
				fmt.Fprintf(b, "    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
				fmt.Fprintf(b, "    public static extern IntPtr %sSet%sNameEmail(\n", opts.CPrefix, sigField)
				fmt.Fprintf(b, "        long handle,\n")
				fmt.Fprintf(b, "        [MarshalAs(UnmanagedType.LPUTF8Str)] string name,\n")
				fmt.Fprintf(b, "        [MarshalAs(UnmanagedType.LPUTF8Str)] string email);\n\n")
			}
		}
		if opts.GoName == "CreateTagOptions" {
			fmt.Fprintf(b, "    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
			fmt.Fprintf(b, "    public static extern IntPtr %sSetTaggerNameEmail(\n", opts.CPrefix)
			fmt.Fprintf(b, "        long handle,\n")
			fmt.Fprintf(b, "        [MarshalAs(UnmanagedType.LPUTF8Str)] string name,\n")
			fmt.Fprintf(b, "        [MarshalAs(UnmanagedType.LPUTF8Str)] string email);\n\n")
		}
		if opts.GoName == "RestoreOptions" {
			fmt.Fprintf(b, "    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
			fmt.Fprintf(b, "    public static extern IntPtr %sAddFile(long handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string path);\n\n", opts.CPrefix)
		}

		fmt.Fprintf(b, "    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]\n")
		fmt.Fprintf(b, "    public static extern void %sFree(long handle);\n\n", opts.CPrefix)
	}
}
