package web_test

import (
	"bytes"
	"io/fs"
	"os"
	"regexp"
	"testing"

	"github.com/adverax/board-demo/web"
)

// tokensHeader is the one line that marks static/tokens.css as derived.
const tokensHeader = "/* derived from meta/design/tokens.css — do not edit (TestTokensCSS_IsVerbatimCopyOfDesignTokens keeps it a byte-for-byte copy below this line) */\n"

// The design tokens have one source of truth, meta/design/tokens.css; the
// embedded copy is that file byte for byte under the marker line.
func TestTokensCSS_IsVerbatimCopyOfDesignTokens(t *testing.T) {
	source, err := os.ReadFile("../meta/design/tokens.css")
	if err != nil {
		t.Fatalf("read the source of truth: %v", err)
	}
	copied, err := fs.ReadFile(web.Page(), "static/tokens.css")
	if err != nil {
		t.Fatal(err)
	}
	if want := append([]byte(tokensHeader), source...); !bytes.Equal(copied, want) {
		t.Fatal("web/static/tokens.css differs from meta/design/tokens.css; regenerate it as the marker line\n" +
			tokensHeader + "followed by meta/design/tokens.css verbatim (see web/README.md)")
	}
}

// Every asset index.html links is embedded, and nothing is loaded from
// another origin (CON-005).
func TestPage_LinksOnlyEmbeddedAssets(t *testing.T) {
	index, err := fs.ReadFile(web.Page(), "index.html")
	if err != nil {
		t.Fatal(err)
	}
	refs := regexp.MustCompile(`(?:href|src)="([^"#][^"]*)"`).FindAllSubmatch(index, -1)
	if len(refs) == 0 {
		t.Fatal("index.html links nothing")
	}
	for _, m := range refs {
		ref := string(m[1])
		if ref == "data:," {
			continue
		}
		if ref[0] != '/' || len(ref) > 1 && ref[1] == '/' {
			t.Errorf("%q is not a same-origin absolute path", ref)
			continue
		}
		if _, err := fs.Stat(web.Page(), ref[1:]); err != nil {
			t.Errorf("%q is not embedded: %v", ref, err)
		}
	}
}
