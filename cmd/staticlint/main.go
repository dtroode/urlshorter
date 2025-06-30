//go:build tools

// Package main provides a static code analyzer for Go projects.
//
// This package combines several types of analyzers:
//   - Standard analyzers from golang.org/x/tools/go/analysis/passes
//   - Analyzers from staticcheck.io
//   - Custom osexit analyzer
//
// Multichecker mechanism:
//
// multichecker.Main() takes a list of analyzers and runs them sequentially
// for each package in the specified directories. Each analyzer receives the AST
// (Abstract Syntax Tree) of the code and can analyze it for potential issues,
// errors, or coding style violations.
//
// Analysis process:
// 1. Parsing Go files into AST
// 2. Sequential execution of each analyzer
// 3. Collection and output of analysis results
// 4. Return non-zero exit code when issues are detected
//
// Usage:
//
//	# Build portable binary from project root
//	go build -o mycheck cmd/staticlint/main.go
//
//	# Run on current project
//	./mycheck cmd/shortener/main.go
//	./mycheck ./...
//
// Portable binary features:
// - Works with any Go project that has a go.mod file
// - Can be distributed and used on any machine with Go installed
// - Automatically detects the target project's module structure
// - No need to modify target project's dependencies
//
// Standard analyzers (golang.org/x/tools/go/analysis/passes):
//
//   - appends: Checks correct usage of append()
//   - asmdecl: Checks correspondence between Go and assembly declarations
//   - assign: Checks correctness of assignments
//   - atomic: Checks correct usage of atomic operations
//   - atomicalign: Checks alignment of atomic fields in structs
//   - bools: Checks simplification of boolean expressions
//   - buildssa: Builds SSA (Static Single Assignment) representation of code
//   - buildtag: Checks correctness of build tags
//   - cgocall: Checks CGO function calls
//   - composite: Checks correctness of composite literals
//   - copylock: Checks copying of structs with locks
//   - ctrlflow: Analyzes control flow
//   - deepequalerrors: Checks correctness of error comparisons
//   - defers: Checks correct usage of defer
//   - directive: Processes analyzer directives
//   - errorsas: Checks correct usage of errors.As
//   - fieldalignment: Checks field alignment in structs
//   - findcall: Finds function calls
//   - framepointer: Checks frame pointer usage
//   - gofix: Applies automatic fixes
//   - hostport: Checks correctness of host:port strings
//   - httpmux: Checks correctness of HTTP multiplexers
//   - httpresponse: Checks HTTP response handling
//   - ifaceassert: Checks interface assertions
//   - inspect: Provides AST inspection
//   - loopclosure: Checks closures in loops
//   - lostcancel: Checks lost context cancellations
//   - nilfunc: Checks nil function calls
//   - nilness: Checks nil values
//   - pkgfact: Collects package facts
//   - printf: Checks correctness of printf functions
//   - reflectvaluecompare: Checks reflect.Value comparisons
//   - shadow: Checks variable shadowing
//   - shift: Checks correctness of bit shifts
//   - sigchanyzer: Checks signal channel usage
//   - slog: Checks correct usage of slog
//   - sortslice: Checks correctness of slice sorting
//   - stdmethods: Checks standard methods
//   - stdversion: Checks standard library versions
//   - stringintconv: Checks string to number conversions
//   - structtag: Checks struct tag correctness
//   - testinggoroutine: Checks goroutines in tests
//   - tests: Checks test correctness
//   - timeformat: Checks time formatting
//   - unmarshal: Checks correctness of unmarshal operations
//   - unreachable: Checks unreachable code
//   - unsafeptr: Checks unsafe.Pointer usage
//   - unusedresult: Checks unused function results
//   - unusedwrite: Checks unused writes
//   - usesgenerics: Checks generic usage
//   - waitgroup: Checks correct usage of WaitGroup
//
// Staticcheck.io analyzers:
//
// All analyzers with "SA" prefix (Security Analysis) are selected along with
// additional analyzers from a predefined list:
//
//   - S1000: Boolean expression simplification
//   - S1001: Loop simplification
//   - S1004: Comparison simplification
//   - S1005: Conditional expression simplification
//   - S1009: Slice operation simplification
//   - S1011: Slice operation simplification
//   - S1016: Slice operation simplification
//   - S1020: Slice operation simplification
//   - S1028: Slice operation simplification
//   - ST1003: Package name checking
//   - ST1005: Error message checking
//   - QF1002: Quick fix - slice operation simplification
//   - QF1003: Quick fix - slice operation simplification
//   - QF1010: Quick fix - slice operation simplification
//
// Custom analyzer:
//
//   - osexit: Checks for os.Exit() usage in main() function of main package.
//     This is important for proper program termination and signal handling.
package main

import (
	"strings"

	"github.com/dtroode/urlshorter/cmd/staticlint/osexit"
	"github.com/kisielk/errcheck/errcheck"

	"github.com/gostaticanalysis/forcetypeassert"
	"github.com/gostaticanalysis/nilerr"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/gofix"
	"golang.org/x/tools/go/analysis/passes/hostport"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"honnef.co/go/tools/staticcheck"
	// Additional analyzers
)

func main() {
	checks := []*analysis.Analyzer{
		appends.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		gofix.Analyzer,
		hostport.Analyzer,
		httpmux.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		slog.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stdversion.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
		waitgroup.Analyzer,
	}

	staticchecks := map[string]bool{
		"S1000":  true,
		"S1001":  true,
		"S1004":  true,
		"S1005":  true,
		"S1009":  true,
		"S1011":  true,
		"S1016":  true,
		"S1020":  true,
		"S1028":  true,
		"ST1003": true,
		"ST1005": true,
		"QF1002": true,
		"QF1003": true,
		"QF1010": true,
	}

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") || staticchecks[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	checks = append(checks, osexit.OsExitAnalyzer)

	// Add errcheck analyzer
	checks = append(checks, errcheck.Analyzer)

	// Add nilerr analyzer
	checks = append(checks, nilerr.Analyzer)

	// Add forcetypeassert analyzer
	checks = append(checks, forcetypeassert.Analyzer)

	multichecker.Main(
		checks...,
	)

}
