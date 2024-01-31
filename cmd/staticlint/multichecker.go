package main

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
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
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
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
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
)

var CheckExitAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "Check for direct calls to os.Exit in the main function",
	Run:  runCheckExitAnalyzer,
}

func runCheckExitAnalyzer(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if call, ok := n.(*ast.CallExpr); ok {
						if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							if id, ok := sel.X.(*ast.Ident); ok && id.Name == "os" {
								if sel.Sel.Name == "Exit" {
									pass.Reportf(call.Pos(), "avoid direct calls to os.Exit in main function")
								}
							}
						}
					}
					return true
				})
			}
			return true
		})
	}
	return nil, nil
}

func main() {
	checks := []*analysis.Analyzer{
		asmdecl.Analyzer,             // Analyzes assembly declarations
		assign.Analyzer,              // Analyzes assignments
		atomic.Analyzer,              // Checks for common mistakes using atomic
		atomicalign.Analyzer,         // Checks for non-atomic bit-wise operations on unaligned values
		bools.Analyzer,               // Simplifies boolean expressions
		buildssa.Analyzer,            // Builds an SSA-form intermediate representation of the program
		buildtag.Analyzer,            // Analyzes build tags
		cgocall.Analyzer,             // Checks for misuse of cgo calls
		composite.Analyzer,           // Checks for unkeyed composite literals
		copylock.Analyzer,            // Detects locks that are mistakenly passed by value
		ctrlflow.Analyzer,            // Checks for issues related to control flow
		deepequalerrors.Analyzer,     // Checks for errors that can be compared with reflect.DeepEqual
		directive.Analyzer,           // Checks for invalid build directives
		errorsas.Analyzer,            // Checks for errors compared to nil
		findcall.Analyzer,            // Finds call sites
		framepointer.Analyzer,        // Analyzes the use of frame pointers in Go programs
		httpresponse.Analyzer,        // Checks for misuse of http.Response
		ifaceassert.Analyzer,         // Checks for suspicious type assertions
		inspect.Analyzer,             // Prints syntax tree, imports, and other details about the program
		loopclosure.Analyzer,         // Checks references to loop variables from within nested functions
		lostcancel.Analyzer,          // Detects loss of cancelation context
		nilfunc.Analyzer,             // Checks for useless calls of functions returning nil
		nilness.Analyzer,             // Checks for dereferences of nil pointers
		pkgfact.Analyzer,             // Checks for calls to fmt.Errorf with a single format specifier
		printf.Analyzer,              // Checks for calls to fmt.Printf that may be incorrect
		reflectvaluecompare.Analyzer, // Checks for equality comparisons between reflect.Value instances
		shadow.Analyzer,              // Checks for shadowed variables
		shift.Analyzer,               // Checks for shifts that equal or exceed the width of the integer
		sigchanyzer.Analyzer,         // Checks for misuse of sync/atomic
		sortslice.Analyzer,           // Suggests simplifications for sorting slices
		stdmethods.Analyzer,          // Checks for misuse of methods defined on built-in types
		stringintconv.Analyzer,       // Suggests simplifications for string-to-int and int-to-string conversions
		structtag.Analyzer,           // Checks for struct tags that contain the word "omitempty"
		testinggoroutine.Analyzer,    // Checks for goroutines that are not waited on
		tests.Analyzer,               // Checks for common mistaken usages of testing functions
		timeformat.Analyzer,          // Checks for calls to time.Sleep with an integer argument
		unmarshal.Analyzer,           // Checks for unmarshal calls where the result is ignored
		unreachable.Analyzer,         // Checks for unreachable code
		unsafeptr.Analyzer,           // Checks for misuse of unsafe.Pointer
	}

	checks = append(checks, CheckExitAnalyzer)

	multichecker.Main(checks...)
}
