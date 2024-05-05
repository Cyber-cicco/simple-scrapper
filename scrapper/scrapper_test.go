package scrapper_test

import (
	"context"
	"os"
	"testing"

	"github.com/Cyber-cicco/simple-scrapper/scrapper"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/html"
)

const DIV_1 = "<div>JavaScript is disabled on your browser.</div>"

const CLASS_1 = `<th class="colOne" scope="col">Constructor and Description</th>`

const ID_1 = `<ul class="navList" id="allclasses_navbar_top">
<li><a href="../../allclasses-noframe.html">All&nbsp;Classes</a></li>
</ul>`

var parser *sitter.Parser
var lang *sitter.Language

func init() {
    parser = sitter.NewParser()
    lang = html.GetLanguage()
    parser.SetLanguage(lang)
}

func TestDOMStructure(t *testing.T) {

	content := initTest(t)

	tree, err := parser.ParseCtx(context.Background(), nil, content)

	if err != nil {
		t.Fatalf("got error %s", err)
	}

	_, err = scrapper.ToDOM(tree.RootNode(), content)

	if err != nil {
		t.Fatalf("got error %s", err)
	}

	_, err = scrapper.ToDOM(tree.RootNode().Child(0), content)

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func TestQuerySelector(t *testing.T) {

	content := initTest(t)

	tree, err := parser.ParseCtx(context.Background(), nil, content)

	if err != nil {
		t.Fatalf("got error %s", err)
	}

	document, err := scrapper.ToDOM(tree.RootNode(), content)

	div1, ok := document.QuerySelector("div")

	if !ok {
		t.Fatalf("QuerySelector returned nil")
	}

	if div1.ToString() != DIV_1 {
		t.Fatalf("Expected %s, got %s", DIV_1, div1.ToString())
	}

	class1, ok := document.QuerySelector(".colOne")

	if !ok {
		t.Fatalf("QuerySelector returned nil")
	}

	if class1.ToString() != CLASS_1 {
		t.Fatalf("Expected %s, got %s", CLASS_1, class1.ToString())
	}

	id1, ok := document.QuerySelector("#allclasses_navbar_top")

	if id1.ToString() != ID_1 {
		t.Fatalf("Expected\n%s, got \n%s", ID_1, id1.ToString())
	}
}

func TestQuerySelectorAll(t *testing.T) {

	content := initTest(t)
	tree, err := parser.ParseCtx(context.Background(), nil, content)

	if err != nil {
		t.Fatalf("got error %s", err)
	}

	document, err := scrapper.ToDOM(tree.RootNode(), content)

    nodes, ok := document.QuerySelectorAll("table")

    if !ok {
        t.Fatalf("Got nil result from QuerySelectorAll")
    }

    if len(nodes) != 4 {
        t.Fatalf("Expected len of 5, got len of %d", len(nodes))
    }
}

func TestInnerText(t *testing.T) {

	content := initTest(t)
	tree, err := parser.ParseCtx(context.Background(), nil, content)

	if err != nil {
		t.Fatalf("got error %s", err)
	}

	document, err := scrapper.ToDOM(tree.RootNode(), content)
	div1, ok := document.QuerySelector("div")

	if !ok {
		t.Fatalf("QuerySelector returned nil")
	}

	text := div1.InnerText()
	expected := "JavaScript is disabled on your browser."

	if string(text) != expected {
		t.Fatalf("Expected %s\n, got %s\n", expected, text)
	}

	expected = "All Classes"
	id1, ok := document.QuerySelector("#allclasses_navbar_top")
	text = id1.InnerText()

	if string(text) != expected {
		t.Fatalf("Expected %s\n, got %s\n", expected, text)
	}
}

func TestAppendingWS(t *testing.T) {
	content := []byte(`<td class="colFirst"><code>static <a href="../../java/lang/ProcessBuilder.Redirect.html" title="class in java.lang">ProcessBuilder.Redirect</a></code></td>`)
	tree, err := parser.ParseCtx(context.Background(), nil, content)

	if err != nil {
		t.Fatalf("got error %s", err)
	}

	document, err := scrapper.ToDOM(tree.RootNode(), content)

    td, ok := document.QuerySelector("td")

    if !ok {
		t.Fatalf("got nil %v", td)
    }

    td.InnerText()
}

func initTest(t *testing.T) []byte {
	path := "ressources/Float.html"
	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("Shouldn't have had error, got %s", err)
	}

	return content
}

