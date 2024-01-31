package main

/*
Package staticlint provides a multichecker for static analysis.

This multichecker includes various static analyzers from golang.org/x/tools/go/analysis/passes,
staticcheck.io, and additional custom analyzers.

Usage:
    go run ./cmd/staticlint [packages]

Analyzers:
- golang.org/x/tools/go/analysis/passes: All analyzers from this package.
- staticcheck.io: All analyzers of class SA and one analyzer from each other class.
- Other classes from staticcheck.io.
- Custom analyzer: exitcheck - Checks for direct calls to os.Exit in the main function.

For more details, refer to the documentation of each analyzer.
*/
