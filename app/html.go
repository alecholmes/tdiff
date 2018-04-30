package app

import (
	"bytes"
	"html/template"
)

var summaryTemplate = template.Must(template.New("").Parse(templateHTML))

func HTML(summary *Summary) ([]byte, error) {
	var buf bytes.Buffer
	if err := summaryTemplate.Execute(&buf, summary); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

var templateHTML = `
<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <title>Go Package Diff</title>
        <style>
            body {
                font-family: monospace;
            }

            .section {
                padding-left: 10px;
            }

            .path {
                padding-left: 1.5em;
                text-indent: -1.5em;
            }
        </style>
    </head>
    <body>
        <h1>Commits</h1>
        <div class="section">
            {{range .Commits}}
                <h3>{{.SHA}}</h3>
                <div class="section">
                    <p><b>{{.Description}}</b></p>
                    {{range .RelevantPackages}}
                        <div class="path">
                            {{range .PathFromRoot}}
                                > {{.}}
                            {{end}}
                        </div>
                    {{end}}
                </div>
            {{end}}
        </div>

        <h1>Packages</h1>
        <div class="section">
            {{range .Packages}}
                <p><b>{{.ImportPath}}</b></p>
                <div class="path">
                    {{range .PathFromRoot}}
                        > {{.}}
                    {{end}}
                </div>
            {{end}}
        </div>

        <h1>Files</h1>
        <div class="section">
            <ul>
                {{range .Files}}
                    <li>{{.}}</li>
                {{end}}
            </ul>
        </div>
    </body>
</html>
`
