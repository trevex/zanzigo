package postgres

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/trevex/zanzigo"
)

var (
	selectQueryTmpl = template.Must(template.New("SelectQuery").Parse(`
{{- $p0 := (index .Placeholders 0) -}}
{{- $p1 := (index .Placeholders 1) -}}
{{- $p2 := (index .Placeholders 2) -}}
{{- $p3 := (index .Placeholders 3) -}}
{{- $KindDirect := .KindDirect -}}
{{- $KindDirectUserset := .KindDirectUserset -}}
{{- $KindIndirect := .KindIndirect -}}
{{- $Brackets := .Brackets -}}

{{- define "relations" -}}
{{- range $i, $relation := . -}}
	{{- if $i }} OR {{ end -}}object_relation='{{ $relation }}'
{{- end -}}
{{- end -}}

{{- range  $id, $rule := .Ruleset }}
	{{ if $id }}UNION ALL{{ end }}
	{{ if $Brackets }}({{ end -}}
	{{- if eq $rule.Kind $KindDirect -}}
	    SELECT {{ $id }} AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples
		 WHERE object_type='{{ $rule.Object }}'
		 AND object_id={{ $p0 }}
		 AND ({{- template "relations" $rule.Relations -}})
		 AND subject_type={{ $p1 }}
		 AND subject_id={{ $p2 }}
		 AND subject_relation={{ $p3 }}
	{{- else if eq $rule.Kind $KindDirectUserset -}}
		SELECT {{ $id }} AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples
		 WHERE object_type='{{ $rule.Object }}'
		 AND object_id={{ $p0 }}
		 AND ({{- template "relations" $rule.Relations -}})
		 AND subject_relation <> ''
	{{- else if eq $rule.Kind $KindIndirect -}}
		SELECT {{ $id }} AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples
		 WHERE object_type='{{ $rule.Object }}'
		 AND object_id={{ $p0 }}
		 AND ({{- template "relations" $rule.Relations -}})
		 AND subject_type='{{ $rule.Subject }}'
	{{- end -}}
	{{- if $Brackets }}){{ end -}}
{{- end }}
	`))
)

func SelectQueryFor(ruleset []zanzigo.InferredRule, brackets bool, placeholders ...string) (string, error) {
	if len(placeholders) == 1 {
		placeholders = []string{placeholders[0], placeholders[0], placeholders[0], placeholders[0]}
	}
	if len(placeholders) != 4 {
		return "", fmt.Errorf("Length of placeholders needs to be 1 or 4")
	}

	var out bytes.Buffer
	err := selectQueryTmpl.Execute(&out, map[string]any{
		"Ruleset":           ruleset,
		"Placeholders":      placeholders,
		"Brackets":          brackets,
		"KindDirect":        zanzigo.KindDirect,
		"KindDirectUserset": zanzigo.KindDirectUserset,
		"KindIndirect":      zanzigo.KindIndirect,
	})

	return strings.Join(strings.Fields(out.String()), " "), err
}

var (
	functionTmpl = template.Must(template.New("SelectQuery").Parse(`
{{- $KindDirect := .KindDirect -}}
{{- $KindDirectUserset := .KindDirectUserset -}}
{{- $KindIndirect := .KindIndirect -}}

CREATE OR REPLACE FUNCTION {{ .FuncName }}(TEXT, TEXT, TEXT, TEXT) RETURNS BOOLEAN LANGUAGE 'plpgsql' AS $$
DECLARE
mt RECORD;
result BOOLEAN;
BEGIN
FOR mt IN
{{ .SelectQuery }} ORDER BY rule_index
LOOP
{{- range  $id, $rule := .Ruleset }}
	{{ if $id }}ELSIF{{ else }}IF{{ end }} mt.rule_index = {{ $id }} THEN
	{{ if eq $rule.Kind $KindDirect }}
		RETURN TRUE;
	{{ else if eq $rule.Kind $KindDirectUserset }}
		EXECUTE FORMAT('SELECT zanzigo_%s_%s($1, $2, $3, $4)', mt.subject_type, mt.subject_relation) USING mt.subject_id, $2, $3, $4 INTO result;
		IF result = TRUE THEN
			RETURN TRUE;
		END IF;
	{{ else if eq $rule.Kind $KindIndirect }}
		{{- range $rule.WithRelationToSubject }}
		SELECT zanzigo_{{ $rule.Subject }}_{{ . }}(mt.subject_id, $2, $3, $4) INTO result;
		IF result = TRUE THEN
			RETURN TRUE;
		END IF;
		{{- end }}
	{{ end }}
{{- end }}
	END IF;
END LOOP;
RETURN FALSE;
END;
$$;`))
)

// TODO: respect maxDepth!
func FunctionFor(funcName string, ruleset []zanzigo.InferredRule) (string, string, error) {
	innerSelect, err := SelectQueryFor(ruleset, true, "$1", "$2", "$3", "$4")
	if err != nil {
		return "", "", err
	}

	var out bytes.Buffer
	err = functionTmpl.Execute(&out, map[string]any{
		"FuncName":          funcName,
		"SelectQuery":       innerSelect,
		"Ruleset":           ruleset,
		"KindDirect":        zanzigo.KindDirect,
		"KindDirectUserset": zanzigo.KindDirectUserset,
		"KindIndirect":      zanzigo.KindIndirect,
	})

	funcDecl := strings.Join(strings.Fields(out.String()), " ")
	funcQuery := "SELECT " + funcName + "($1, $2, $3, $4)"
	return funcDecl, funcQuery, err
}
