/*
Package main provides complex multichecker, which consists of:
  - exitcheck.Analyzer 				checks calling os.Exit()
  - shift.Analyzer:					checks for shifts that exceed the width of an integer.
  - printf.Analyzer					checks consistency of Printf format strings and arguments
  - shadow.Analyzer					checks for possible unintended shadowing of variables
  - structtag.Analyzer				checks struct field tags are well formed.
  - staticcheck						all SA*, S1*, ST1* checks from static checks (staticcheck.io)
  - nakedret.NakedReturnAnalyzer 	checks naked returns in functions greater than a specified function length (github.com/alexkohler/nakedret)
  - bidichk.NewAnalyzer				checks dangerous unicode character sequences in Go source files (https://github.com/breml/bidichk)

Using:
 1. build package:
    $> go build -o staticlint
 2. exec multichecker
    $> ./staticlint ./cmd/main.go
 3. for information type
    $> ./staticlint help
*/
package main
