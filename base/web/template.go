package web

var share_template = map[string]interface{}{
	"shy/story": `{{}}`,
	"shy/chain": `<!DOCTYPE html>
<head>
<meta charset="utf-8">
<link rel="stylesheet" text="text/css" href="/style.css">
</head>
<body>
<fieldset>
	{{.Append "type"}}
	{{.Append "text"}}
</fieldset>
</body>
`,
}
