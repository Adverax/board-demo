package board_test

import (
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// ADR-0004/R1: package board does not import net/http or any other HTTP
// concern. The ADR notes this boundary erodes silently without a check.
func TestBoard_DoesNotImportHTTP(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if strings.HasPrefix(path, "net/http") || strings.HasSuffix(path, "/httpapi") {
				t.Errorf("%s imports %s", name, path)
			}
		}
	}
}
