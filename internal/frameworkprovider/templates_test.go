package frameworkprovider

import (
	"bytes"
	"embed"
	"html/template"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func executeTemplate(t *testing.T, name string, data any) string {
	tmpl := getTemplates(t, name)
	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, data))
	return buf.String()
}

//go:embed test_data/templates
var templatesFS embed.FS

var templates = struct {
	once sync.Once
	tmpl *template.Template
}{
	once: sync.Once{},
	tmpl: nil,
}

func getTemplates(t *testing.T, name string) *template.Template {
	templates.once.Do(func() {
		var err error
		templates.tmpl, err = template.
			New("").
			Funcs(template.FuncMap{
				"valueString": typesValueToString,
			}).
			ParseFS(templatesFS, "test_data/templates/*.tmpl")
		require.NoError(t, err)
	})
	tmpl := templates.tmpl.Lookup(name)
	require.NotNil(t, tmpl)
	return tmpl
}

type valueStringer interface {
	ValueString() string
}

func typesValueToString(vs valueStringer) string {
	return vs.ValueString()
}
