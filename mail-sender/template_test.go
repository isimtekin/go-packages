package mailsender

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderHTMLTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateStr  string
		data         interface{}
		wantContains string
		wantErr      bool
		checkError   func(t *testing.T, err error)
	}{
		{
			name:         "simple template",
			templateStr:  "<h1>Hello {{.Name}}</h1>",
			data:         map[string]string{"Name": "World"},
			wantContains: "<h1>Hello World</h1>",
			wantErr:      false,
		},
		{
			name: "complex template",
			templateStr: `
				<html>
					<body>
						<h1>Hello {{.Name}}</h1>
						<p>You have {{.Count}} new messages.</p>
					</body>
				</html>
			`,
			data: map[string]interface{}{
				"Name":  "John",
				"Count": 5,
			},
			wantContains: "Hello John",
			wantErr:      false,
		},
		{
			name:        "invalid template syntax",
			templateStr: "<h1>Hello {{.Name</h1>",
			data:        map[string]string{"Name": "World"},
			wantErr:     true,
			checkError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "failed to parse HTML template")
			},
		},
		{
			name:         "template with range",
			templateStr:  "<ul>{{range .Items}}<li>{{.}}</li>{{end}}</ul>",
			data:         map[string][]string{"Items": {"Item1", "Item2", "Item3"}},
			wantContains: "<li>Item1</li>",
			wantErr:      false,
		},
		{
			name:         "template with missing field (returns empty)",
			templateStr:  "<p>{{.MissingField}}</p>",
			data:         map[string]string{"Name": "World"},
			wantContains: "<p></p>",
			wantErr:      false,
		},
		{
			name:         "empty template",
			templateStr:  "",
			data:         map[string]string{},
			wantContains: "",
			wantErr:      false,
		},
		{
			name:         "nil data",
			templateStr:  "<p>Static HTML</p>",
			data:         nil,
			wantContains: "<p>Static HTML</p>",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderHTMLTemplate(tt.templateStr, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkError != nil {
					tt.checkError(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Contains(t, result, tt.wantContains)
			}
		})
	}
}

func TestRenderTextTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateStr  string
		data         interface{}
		wantContains string
		wantErr      bool
		checkError   func(t *testing.T, err error)
	}{
		{
			name:         "simple template",
			templateStr:  "Hello {{.Name}}",
			data:         map[string]string{"Name": "World"},
			wantContains: "Hello World",
			wantErr:      false,
		},
		{
			name: "multiline template",
			templateStr: `
Hello {{.Name}},

You have {{.Count}} new messages.

Best regards,
The Team
			`,
			data: map[string]interface{}{
				"Name":  "John",
				"Count": 5,
			},
			wantContains: "Hello John",
			wantErr:      false,
		},
		{
			name:        "invalid template syntax",
			templateStr: "Hello {{.Name",
			data:        map[string]string{"Name": "World"},
			wantErr:     true,
			checkError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "failed to parse text template")
			},
		},
		{
			name: "template with conditionals",
			templateStr: `Hello {{.Name}}
{{if .Premium}}You are a premium member!{{else}}Upgrade to premium today!{{end}}`,
			data: map[string]interface{}{
				"Name":    "John",
				"Premium": true,
			},
			wantContains: "You are a premium member!",
			wantErr:      false,
		},
		{
			name:         "template with missing field (returns empty)",
			templateStr:  "Hello {{.MissingField}}",
			data:         map[string]string{"Name": "World"},
			wantContains: "Hello ",
			wantErr:      false,
		},
		{
			name:         "empty template",
			templateStr:  "",
			data:         map[string]string{},
			wantContains: "",
			wantErr:      false,
		},
		{
			name:         "nil data",
			templateStr:  "Hello World",
			data:         nil,
			wantContains: "Hello World",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderTextTemplate(tt.templateStr, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkError != nil {
					tt.checkError(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Contains(t, result, tt.wantContains)
			}
		})
	}
}
