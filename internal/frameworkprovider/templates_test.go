package frameworkprovider

import (
	"bytes"
	"embed"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"text/template"

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
				"valueString":      typesValueToStringTplFunc,
				"hasField":         hasFieldTplFunc,
				"renderMetricSpec": renderMetricSpecTplFunc,
				"isNull":           isNullTplFunc,
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

func typesValueToStringTplFunc(vs valueStringer) string {
	return vs.ValueString()
}

func hasFieldTplFunc(name string, v interface{}) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return false
	}
	return rv.FieldByName(name).IsValid()
}

func renderMetricSpecTplFunc(metricSpec interface{}) string {
	rv := reflect.ValueOf(metricSpec)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ""
	}
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		structField := rt.Field(i)
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			return renderMetricTypeFields(structField.Tag.Get("tfsdk"), field.Interface())
		}
	}
	return ""
}

func renderMetricTypeFields(blockName string, metricModel interface{}) string {
	rv := reflect.ValueOf(metricModel)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ""
	}
	var fields []string
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		structField := rt.Field(i)
		if method := field.MethodByName("IsNull"); method.IsValid() && method.Call(nil)[0].Bool() {
			continue
		}
		fields = append(fields, fmt.Sprintf(`%s = %v`, structField.Tag.Get("tfsdk"), field.Interface()))
	}
	if len(fields) == 0 {
		return ""
	}
	b := strings.Builder{}
	b.WriteString("\n")
	b.WriteString(strings.Repeat(" ", 8))
	b.WriteString(blockName)
	b.WriteString(" {\n")
	for _, field := range fields {
		b.WriteString(strings.Repeat(" ", 10) + field + "\n")
	}
	b.WriteString(strings.Repeat(" ", 8))
	b.WriteString("}")
	return b.String()
}

func isNullTplFunc(given interface{}) bool {
	rv := reflect.ValueOf(given)
	method := rv.MethodByName("IsNull")
	if method.IsValid() && method.Call(nil)[0].Bool() {
		return true
	}
	return false
}
