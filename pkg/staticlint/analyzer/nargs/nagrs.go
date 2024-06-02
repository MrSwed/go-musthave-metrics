package nargs

import (
	"go/token"
	"strconv"
	"strings"

	"github.com/alexkohler/nargs"
	"golang.org/x/tools/go/analysis"
)

const errNACheck = "Find unused arguments in function declarations"

// NArgsAnalyzer
// WARN! it is not worked yet, tried to use github.com/alexkohler/nargs
var NArgsAnalyzer = &analysis.Analyzer{
	Name: "nargs",
	Doc:  errNACheck,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// todo: that has once public method,  it is possible to use it ?
		results, _, err := nargs.CheckForUnusedFunctionArgs([]string{file.Name.String()}, nargs.Flags{
			IncludeTests:        true,
			SetExitStatus:       true,
			IncludeNamedReturns: true,
			IncludeReceivers:    true,
		})
		if err != nil {
			pass.Reportf(file.Pos(), "nargs: %s", err)
		}
		// if exitWithStatus {
		// 	pass.Reportf(file.Pos(), "nargs: exit withstatus")
		// }

		// if results != nil {
		// 	pass.Reportf(file.Pos(), "nargs: %s", results)
		// }
		for i := 0; i < len(results); i++ {
			a := strings.SplitN(results[i], ":", 2)
			var r int
			if len(a) == 2 {
				a1 := strings.SplitN(a[1], " ", 2)
				r, err = strconv.Atoi(a1[0])
			}
			pass.Reportf(token.Pos(r), "nargs: %s", results[i])

		}
	}
	return nil, nil
}
