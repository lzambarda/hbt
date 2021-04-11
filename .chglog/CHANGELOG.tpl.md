# CHANGELOG

{{ if .Versions -}}
<a name="unreleased"></a>
## [Unreleased]
{{ if .Unreleased.CommitGroups -}}
{{ range .Unreleased.CommitGroups }}
### {{ .Title }}
{{ range .Commits -}}
{{ if not (contains .Subject "<code>") -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ if .Subject }}{{ .Subject }}{{ else }}{{ .Header }}{{ end }}
{{ end -}}
{{ end -}}
{{ end }}
{{ else }}
{{ range .Unreleased.Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ if .Subject }}{{ .Subject }}{{ else }}{{ .Header }}{{ end }}
{{ end }}
{{ end -}}

{{- if .Unreleased.NoteGroups -}}
{{ range .Unreleased.NoteGroups -}}
### {{ .Title }}
{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}

{{ range .Versions }}
<a name="{{ .Tag.Name }}"></a>
## {{ if .Tag.Previous }}[{{ .Tag.Name }}]{{ else }}{{ .Tag.Name }}{{ end }} - {{ datetime "2006-01-02" .Tag.Date }}
{{ if .CommitGroups -}}
{{ range .CommitGroups }}
### {{ .Title }}
{{ range .Commits -}}
{{ if not (contains .Subject "<code>") -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ if .Subject }}{{ .Subject }}{{ else }}{{ .Header }}{{ end }}
{{ end -}}
{{ end -}}
{{ end }}
{{ else }}
{{ range .Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ if .Subject }}{{ .Subject }}{{ else }}{{ .Header }}{{ end }}
{{ end }}
{{ end -}}

{{- if .NoteGroups -}}
{{ range .NoteGroups -}}
### {{ .Title }}
{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}