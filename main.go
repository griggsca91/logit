package main

import (
	"fmt"
	"io"

	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
)

const logFunction = `
func l(name string) {
  fmt.Printf("function %s\n", name)
}
`

func logFunctionCall(name string) string {
	return fmt.Sprintf(" l(\"%s\")\n", name)
}

func getNewLinePosition(sourceCode []byte, start int) (int, error) {

	for i := start; i < len(sourceCode); i++ {
		if sourceCode[i] == '\n' {
			return i, nil
		}
	}

	return 0, fmt.Errorf("no newline after position %v", start)
}

func injectLogs(sourceCode []byte) ([]byte, error) {
	fset := token.NewFileSet()
	p, err := parser.ParseFile(fset, "main.go", sourceCode, 0)
	if err != nil {
		return nil, err
	}
	slices.Reverse(p.Decls)
	// Print the location and kind of each declaration in f.
	for _, decl := range p.Decls {
		// Get the filename, line, and column back via the file set.
		// We get both the relative and absolute position.
		// The relative position is relative to the last line directive.
		// The absolute position is the exact position in the source.
		pos := decl.Pos()

		// Either a FuncDecl or GenDecl, since we exit on error.
		funk, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		end, err := getNewLinePosition(sourceCode, int(pos-1))
		if err != nil {
			return nil, err
		}
		fmt.Println(string(sourceCode[pos-1 : end]))

		sourceCode = append(sourceCode[0:end+1],
			append(
				[]byte(logFunctionCall(funk.Name.Name)),
				sourceCode[end:]...,
			)...,
		)
	}

	sourceCode = append(sourceCode, []byte(logFunction)...)
	return sourceCode, nil
}

func main() {
	if len(os.Args) <= 1 {
		panic("not enough args")
	}
	f, err := os.OpenFile(os.Args[1], os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	sourceCode, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	sourceCode, err = injectLogs(sourceCode)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("main_new.go", sourceCode, 0)
	if err != nil {
		panic(err)
	}

	fmt.Println(os.Args)
}

func testFunction() {

}
