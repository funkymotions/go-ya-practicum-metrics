package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const filename = "reset.gen.go"

// Tool scans packages under the provided root (default: current dir), finds
// structs annotated with `generate:reset`, and generates Reset methods for
// them in reset.gen.go within the same package. The generated Reset follows
// the rules: primitives to zero, slices to [:0], maps cleared, pointers reset
// recursively, and nested structs call their Reset if present.
func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <root-path>\n", os.Args[0])
		return
	}

	rootDir := os.Args[1]
	if rootDir == "" {
		rootDir = "."
	}

	fmt.Printf("Generating Reset() from root path: %s\n", rootDir)
	files, err := findGoFiles(rootDir)
	if err != nil {
		fmt.Printf("Error finding Go files: %v\n", err)

		return
	}

	markedFiles, err := getMarkedFiles(files)
	if err != nil {
		fmt.Printf("no files with // generate:reset found\n")
		return
	}

	for _, fileInfo := range markedFiles {
		fmt.Printf("Processing file: %s\n", fileInfo.path)
		handleGenerateResetForFile(fileInfo)
	}
}

func findGoFiles(rootDir string) ([]string, error) {
	goFiles := make([]string, 0)
	filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})

	if len(goFiles) == 0 {
		return nil, fmt.Errorf("no Go files found for root: %s", rootDir)
	}

	return goFiles, nil
}

type fileInfo struct {
	path    string
	ast     *ast.File
	fileSet *token.FileSet
}

func getMarkedFiles(files []string) ([]fileInfo, error) {
	var result []fileInfo
	for _, path := range files {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing file %s: %v\n", path, err)
			continue
		}
		hasGanerate := false
		for _, commentGroup := range file.Comments {
			for _, comment := range commentGroup.List {
				if strings.Contains(comment.Text, "// generate:reset") {
					hasGanerate = true
					break
				}
			}
			if hasGanerate {
				break
			}
		}

		if hasGanerate {
			result = append(result, fileInfo{path: path, ast: file, fileSet: fset})
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no files with // generate:reset found")
	}

	return result, nil
}

func handleGenerateResetForFile(fileInfo fileInfo) {
	p := filepath.Dir(fileInfo.path) + "/" + filename
	file, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", p, err)
		return
	}
	defer file.Close()
	packageLine := fmt.Sprintf("package %s\n", fileInfo.ast.Name.Name)
	if _, err := file.WriteString(packageLine); err != nil {
		fmt.Printf("Error writing to file %s: %v\n", p, err)
	}

	ast.Inspect(fileInfo.ast, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		// if not a structure, skip it
		if _, ok := spec.Type.(*ast.StructType); !ok {
			return false
		}

		// extract comments group
		cg := extractDocComments(spec, fileInfo.ast)
		if hasGenerateResetComment(cg) {
			methodCode := generateResetMethod(fileInfo.fileSet, fileInfo.ast, spec)
			f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("Error opening file %s: %v\n", p, err)
				return false
			}
			defer f.Close()
			if _, err := f.WriteString("\n" + methodCode); err != nil {
				fmt.Printf("Error writing to file %s: %v\n", p, err)
			}
		}

		return false
	})
}

func extractDocComments(ts *ast.TypeSpec, file *ast.File) *ast.CommentGroup {
	for _, decl := range file.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range gd.Specs {
				if spec == ts {
					return gd.Doc
				}
			}
		}
	}

	return nil
}

func hasGenerateResetComment(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, comment := range cg.List {
		if strings.Contains(comment.Text, "generate:reset") {
			return true
		}
	}

	return false
}

func generateResetMethod(fset *token.FileSet, file *ast.File, ts *ast.TypeSpec) string {
	sturctName := ts.Name.Name
	out := strings.Builder{}
	nl := []byte{'\n'}
	signature := fmt.Sprintf(`func (x *%s) Reset() {`, sturctName)
	out.Write([]byte(signature))
	out.Write(nl)
	for _, field := range ts.Type.(*ast.StructType).Fields.List {
		for _, name := range field.Names {
			resetLine := getResetDirectiveByField(name, field.Type, fset, file) + "\n"
			out.Write([]byte(resetLine))
		}
	}

	return out.String() + "}\n"
}

func getResetDirectiveByField(fieldName *ast.Ident, t ast.Expr, _ *token.FileSet, file *ast.File) string {
	switch tt := t.(type) {
	case *ast.Ident:
		if stmt := resetPrimitive(fieldName, tt); stmt != "" {
			return stmt
		}
		return resetStruct(fieldName, tt.Name, file)
	case *ast.ArrayType:
		if tt.Len == nil {
			return resetSlice(fieldName)
		}
		return ""
	case *ast.MapType:
		return resetMap(fieldName)
	case *ast.StarExpr:
		return resetPointer(fieldName, tt.X, file)
	}

	return ""
}

func resetPrimitive(fieldName *ast.Ident, ident *ast.Ident) string {
	switch ident.Name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint16", "uint32", "uint64",
		"float32", "float64", "complex64", "complex128":
		return fmt.Sprintf("\tx.%s = 0", fieldName.Name)
	case "string":
		return fmt.Sprintf("\tx.%s = \"\"", fieldName.Name)
	case "bool":
		return fmt.Sprintf("\tx.%s = false", fieldName.Name)
	default:
		return ""
	}
}

func resetPointer(fieldName *ast.Ident, ident ast.Expr, file *ast.File) string {
	if ident == nil {
		return ""
	}
	ptrExpr := "x." + fieldName.Name // pointer variable
	valExpr := "*" + ptrExpr         // dereferenced value
	var body string
	switch et := ident.(type) {
	case *ast.Ident:
		if stmt := resetPrimitiveExpr(valExpr, et); stmt != "" {
			body = stmt
		} else if et.Name != "" && hasResetMethod(et.Name, file) {
			body = fmt.Sprintf("\t%s.Reset()", ptrExpr)
		}
	case *ast.ArrayType:
		if et.Len == nil { // slice
			body = fmt.Sprintf("\t%s = %s[:0]", valExpr, valExpr)
		}
	case *ast.MapType:
		body = fmt.Sprintf("\tclear(%s)", valExpr)
	}

	if body == "" {
		body = fmt.Sprintf("\t// TODO: reset for pointer field %s (no Reset() method declared)", fieldName.Name)
	}

	return fmt.Sprintf("\tif %s != nil {\n\t%s\n\t}", ptrExpr, body)
}

func resetPrimitiveExpr(expr string, ident *ast.Ident) string {
	switch ident.Name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint16", "uint32", "uint64",
		"float32", "float64", "complex64", "complex128":
		return fmt.Sprintf("\t%s = 0", expr)
	case "string":
		return fmt.Sprintf("\t%s = \"\"", expr)
	case "bool":
		return fmt.Sprintf("\t%s = false", expr)
	default:
		return ""
	}
}

func resetSlice(fieldName *ast.Ident) string {
	return fmt.Sprintf("\tx.%s = x.%s[:0]", fieldName.Name, fieldName.Name)
}

func resetMap(fieldName *ast.Ident) string {
	return fmt.Sprintf("\tif x.%s != nil {\n\t\tclear(x.%s)\n\t}", fieldName.Name, fieldName.Name)
}

func resetStruct(fieldName *ast.Ident, typeName string, file *ast.File) string {
	if typeName != "" && hasResetMethod(typeName, file) {
		return fmt.Sprintf("\tx.%s.Reset()", fieldName.Name)
	}

	return fmt.Sprintf("\t// TODO: reset for struct field %s (%s) without Reset()", fieldName.Name, typeName)
}

func hasResetMethod(typeName string, file *ast.File) bool {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name == nil || fn.Name.Name != "Reset" {
			continue
		}
		for _, recv := range fn.Recv.List {
			switch rt := recv.Type.(type) {
			case *ast.Ident:
				if rt.Name == typeName {
					return true
				}
			case *ast.StarExpr:
				if id, ok := rt.X.(*ast.Ident); ok && id.Name == typeName {
					return true
				}
			}
		}
	}

	return false
}
