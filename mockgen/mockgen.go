package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/lvan100/gomock/gomock"
)

var flagVar struct {
	outputFile     string
	mockInterfaces string
}

func init() {
	flag.StringVar(&flagVar.outputFile, "o", "", "output file")
	flag.StringVar(&flagVar.outputFile, "output", "", "output file")
	flag.StringVar(&flagVar.mockInterfaces, "t", "", "mock interfaces")
	flag.StringVar(&flagVar.mockInterfaces, "interfaces", "", "mock interfaces")
}

func main() {
	flag.Parse()
	run(runParam{
		sourceDir:      ".",
		outputFile:     flagVar.outputFile,
		mockInterfaces: flagVar.mockInterfaces,
	})
}

type runParam struct {
	sourceDir      string
	outputFile     string
	mockInterfaces string
}

func run(param runParam) {
	ctx := scanContext{
		outputFile:        param.outputFile,
		includeInterfaces: make(map[string]struct{}),
		excludeInterfaces: make(map[string]struct{}),
	}

	ctx.parse(param.mockInterfaces)
	arr := scanDir(param.sourceDir, ctx)

	imports := make(map[string]struct{})
	for _, m := range arr {
		for _, pkgPath := range m.Imports {
			imports[pkgPath] = struct{}{}
		}
	}

	var sb strings.Builder
	sb.WriteString("package testdata\n")
	sb.WriteString("\n\nimport (\n")
	sb.WriteString("\t\"reflect\"\n")
	for pkgPath := range imports {
		sb.WriteString(fmt.Sprintf("\t\"%s\"\n", pkgPath))
	}
	sb.WriteString("\t\"github.com/lvan100/gomock/gomock\"\n")
	sb.WriteString(")\n")

	for _, m := range arr {
		mockInterface(&sb, m)
	}

	b, err := format.Source([]byte(sb.String()))
	if err != nil {
		panic(err)
	}

	if param.outputFile == "" {
		if _, err = os.Stdout.Write(b); err != nil {
			panic(err)
		}
		return
	}

	outputFile := filepath.Join(param.sourceDir, param.outputFile)
	err = os.WriteFile(outputFile, b, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

type scanContext struct {
	outputFile        string
	includeInterfaces map[string]struct{}
	excludeInterfaces map[string]struct{}
}

func (ctx *scanContext) parse(mockInterfaces string) {
	if c := mockInterfaces[0]; c == '\'' || c == '"' {
		mockInterfaces = mockInterfaces[1 : len(mockInterfaces)-1]
	}
	ss := strings.Split(mockInterfaces, ",")
	for _, s := range ss {
		if s[0] == '!' {
			ctx.excludeInterfaces[s[1:]] = struct{}{}
		} else {
			ctx.includeInterfaces[s] = struct{}{}
		}
	}
}

func (ctx *scanContext) mock(name string) bool {
	if len(ctx.includeInterfaces) > 0 {
		_, ok := ctx.includeInterfaces[name]
		return ok
	}
	_, ok := ctx.excludeInterfaces[name]
	return !ok
}

type MockInterface struct {
	Name    string
	Type    *ast.InterfaceType
	File    string
	Imports map[string]string
}

func scanDir(dir string, ctx scanContext) []MockInterface {
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	var ret []MockInterface
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == ctx.outputFile {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		arr := scanFile(ctx, filepath.Join(dir, entry.Name()))
		ret = append(ret, arr...)
	}
	return ret
}

func scanFile(ctx scanContext, file string) []MockInterface {
	mode := parser.AllErrors
	node, err := parser.ParseFile(token.NewFileSet(), file, nil, mode)
	if err != nil {
		panic(err)
	}

	imports := make(map[string]string)
	for _, spec := range node.Imports {
		pkgPath := spec.Path.Value
		pkgPath = pkgPath[1 : len(pkgPath)-1]
		var pkgName string
		if spec.Name != nil {
			pkgName = spec.Name.Name
		} else {
			ss := strings.Split(pkgPath, "/")
			pkgName = ss[len(ss)-1]
		}
		imports[pkgName] = pkgPath
	}

	var ret []MockInterface
	for _, decl := range node.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok || d.Tok != token.TYPE {
			continue
		}
		for _, spec := range d.Specs {
			s := spec.(*ast.TypeSpec)
			t, ok := s.Type.(*ast.InterfaceType)
			if !ok || len(t.Methods.List) == 0 {
				continue
			}
			name := s.Name.String()
			if !ctx.mock(name) {
				continue
			}
			ret = append(ret, MockInterface{
				Name:    name,
				Type:    t,
				File:    file,
				Imports: imports,
			})
		}
	}
	return ret
}

func mockInterface(sb *strings.Builder, mi MockInterface) {
	sb.WriteString(fmt.Sprintf("\ntype %sMockImpl struct {\n\tr *gomock.Manager\n}\n", mi.Name))
	sb.WriteString(fmt.Sprintf("\nfunc New%sMockImpl(r *gomock.Manager) *%sMockImpl {", mi.Name, mi.Name))
	sb.WriteString(fmt.Sprintf("\n\treturn &%sMockImpl{r}", mi.Name))
	sb.WriteString("\n}\n")
	for _, method := range mi.Type.Methods.List {
		methodName := method.Names[0].Name
		ft := method.Type.(*ast.FuncType)
		paramCount := len(ft.Params.List)
		if paramCount > gomock.MaxParamCount {
			panic(fmt.Sprintf("have more than %d parameters", gomock.MaxParamCount))
		}
		resultCount := len(ft.Results.List)
		if resultCount > gomock.MaxResultCount {
			panic(fmt.Sprintf("have more than %d results", gomock.MaxResultCount))
		}
		{
			sb.WriteString(fmt.Sprintf("\nfunc (impl *%sMockImpl) %s(", mi.Name, methodName))
			for i, param := range ft.Params.List {
				sb.WriteString(param.Names[0].Name)
				sb.WriteString(" ")
				_ = printer.Fprint(sb, token.NewFileSet(), param.Type)
				if i < len(ft.Params.List)-1 {
					sb.WriteString(", ")
				}
			}
			sb.WriteString(") (")
			for i, result := range ft.Results.List {
				_ = printer.Fprint(sb, token.NewFileSet(), result.Type)
				if i < len(ft.Results.List)-1 {
					sb.WriteString(", ")
				}
			}
			sb.WriteString(") {")
			sb.WriteString(fmt.Sprintf("\n\tt := reflect.TypeFor[%sMockImpl]()", mi.Name))
			sb.WriteString(fmt.Sprintf("\n\tif ret, ok := gomock.Invoke(impl.r, t, \"%s\", ", methodName))
			for i, param := range ft.Params.List {
				sb.WriteString(param.Names[0].Name)
				if i < len(ft.Params.List)-1 {
					sb.WriteString(", ")
				}
			}
			sb.WriteString("); ok {")
			sb.WriteString(fmt.Sprintf("\n\t\treturn gomock.Unbox%d[", resultCount))
			for i, result := range ft.Results.List {
				_ = printer.Fprint(sb, token.NewFileSet(), result.Type)
				if i < len(ft.Results.List)-1 {
					sb.WriteString(", ")
				}
			}
			sb.WriteString("](ret)")
			sb.WriteString("\n\t}")
			sb.WriteString("\n\tpanic(\"no mock code matched\")")
			sb.WriteString("\n}")
		}
		sb.WriteString("\n")
		{
			sb.WriteString(fmt.Sprintf("\nfunc (impl *%sMockImpl) Mock%s(", mi.Name, methodName))
			sb.WriteString(fmt.Sprintf(") *gomock.Mocker%d%d[", paramCount, resultCount))
			for i, param := range ft.Params.List {
				_ = printer.Fprint(sb, token.NewFileSet(), param.Type)
				if i < len(ft.Params.List)-1 {
					sb.WriteString(", ")
				}
			}
			sb.WriteString(", ")
			for i, result := range ft.Results.List {
				_ = printer.Fprint(sb, token.NewFileSet(), result.Type)
				if i < len(ft.Results.List)-1 {
					sb.WriteString(", ")
				}
			}
			sb.WriteString("] {")
			sb.WriteString(fmt.Sprintf("\n\tt := reflect.TypeFor[%sMockImpl]()", mi.Name))
			sb.WriteString(fmt.Sprintf("\n\treturn gomock.NewMocker%d%d[", paramCount, resultCount))
			for i, param := range ft.Params.List {
				_ = printer.Fprint(sb, token.NewFileSet(), param.Type)
				if i < len(ft.Params.List)-1 {
					sb.WriteString(", ")
				}
			}
			sb.WriteString(", ")
			for i, result := range ft.Results.List {
				_ = printer.Fprint(sb, token.NewFileSet(), result.Type)
				if i < len(ft.Results.List)-1 {
					sb.WriteString(", ")
				}
			}
			sb.WriteString(fmt.Sprintf("](impl.r, t, \"%s\")", methodName))
			sb.WriteString("\n}")
		}
	}
}
