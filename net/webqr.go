package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":1718", "http service address")

var templ = template.Must(template.New("qr").Parse(templateStr))

func main() {
	flag.Parse()
	http.HandleFunc("/", QR)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func QR(w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, r.FormValue("s"))
}

const templateStr = `
<html>
<head>
<title>QR link generator</title>
</head>
<body>
{{if .}}
<img src="http://chart.apis.google.com/chart?chs=300x300&cht=qr&choe=UTF-8&chl={{.}}" />
<br>
{{.}}
<br>
<br>
{{end}}
<form action="/" name=f method="GET">
<input maxLength=1024 size=70 name = s value = "" title="Text to QR encode">
<input type=submit value= "Show QR" name=qr>
</form>
</body>
</html>
`
