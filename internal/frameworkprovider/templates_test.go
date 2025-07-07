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
				"escapeString":     escapeStringTplFunc,
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
		// Check if field is a slice with at least one element
		if field.Kind() == reflect.Slice && field.Len() > 0 {
			// Get the first element of the slice
			firstElem := field.Index(0)
			return renderMetricTypeFields(structField.Tag.Get("tfsdk"), firstElem.Interface(), 8)
		}
	}
	return ""
}

func renderMetricTypeFields(blockName string, metricModel interface{}, baseIndent int) string {
	rv := reflect.ValueOf(metricModel)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return rv.String()
	}
	var fields []string
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		structField := rt.Field(i)
		if field.IsZero() {
			continue
		}
		if method := field.MethodByName("IsNull"); method.IsValid() && method.Call(nil)[0].Bool() {
			continue
		}
		fieldName := structField.Tag.Get("tfsdk")
		var fieldValue string
		switch field.Kind() {
		case reflect.String:
			fieldValue = fmt.Sprintf(`"%s"`, escapeStringTplFunc(field.String()))
			fields = append(fields, fmt.Sprintf(`%s = %s`, fieldName, fieldValue))
		case reflect.Slice:
			if field.Len() == 0 {
				continue
			}
			firstElem := field.Index(0)
			fieldValue = renderMetricTypeFields(fieldName, firstElem.Interface(), baseIndent+2)
			fields = append(fields, fieldValue)
		default:
			fieldValue = fmt.Sprintf(`%v`, field.Interface())
			fields = append(fields, fmt.Sprintf(`%s = %s`, fieldName, fieldValue))
		}
	}
	if len(fields) == 0 {
		return ""
	}
	b := strings.Builder{}
	b.WriteString("\n")
	b.WriteString(strings.Repeat(" ", baseIndent))
	b.WriteString(blockName)
	b.WriteString(" {\n")
	for _, field := range fields {
		if strings.HasPrefix(field, "\n") {
			field = strings.TrimPrefix(field, "\n")
			b.WriteString(field + "\n")
		} else {
			b.WriteString(strings.Repeat(" ", baseIndent+2) + field + "\n")
		}
	}
	b.WriteString(strings.Repeat(" ", baseIndent))
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

func escapeStringTplFunc(s string) string {
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "\n", `\n`)
	return strings.ReplaceAll(s, `"`, `\"`)
}
