// Package osexit provides an analyzer for checking os.Exit() usage in main() function.
//
// This analyzer is part of the static code analyzer and checks for correct usage
// of os.Exit() in main packages. Using os.Exit() in main() function can lead to
// problems with:
//   - Proper program termination
//   - Signal handling (SIGTERM, SIGINT)
//   - Execution of defer functions
//   - Logging of program shutdown
//
// The analyzer only checks files in the main package and issues a warning
// when os.Exit() call is detected.
//
// Example of problematic code:
//
//	package main
//
//	import "os"
//
//	func main() {
//		// ... code ...
//		os.Exit(1) // Warning: os call detected
//	}
//
// Recommended alternatives:
//   - Return non-zero code from main()
//   - Use log.Fatal() for logging and termination
//   - Proper error handling with return codes
package osexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer is a structure that sets up custom analyzer.
// It specifies name, documentation and function to run.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check for os.Exit() in func main() of package main",
	Run:  run,
}

// run is the main function of the analyzer that processes each file in the package.
// It traverses the AST looking for os.Exit() calls and reports them as issues.
//
// The function observes only main() function in main package.
func run(pass *analysis.Pass) (any, error) {
	inspect := func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Name == "os" && sel.Sel.Name == "Exit" {
			pass.Reportf(ident.Pos(), "os call detected")
		}
		return true
	}

	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		for _, d := range file.Decls {
			funcDecl, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if funcDecl.Name.Name == "main" {
				ast.Inspect(funcDecl, inspect)
				break
			}
		}
		break
	}
	return nil, nil
}
