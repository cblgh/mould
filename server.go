package main

import (
	"fmt"
	"errors"
	"os"
	"syscall"
	"net/http"
	"mould/myform"
	crand "crypto/rand"
	"math/big"
	"html/template"
	"strings"
	"encoding/json"
	_ "embed"
)

type RequestHandler struct {
}

func (h RequestHandler) ErrorRoute(res http.ResponseWriter, req *http.Request) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
}

//go:embed index-template.html
var htmlContents string

var responses map[string]string

// used for generating a random identifier
const characterSet = "abcdedfghijklmnopqrstABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const pwlength = 20

func generateResponseIdentifier() string {
	var identifier strings.Builder
	const maxChar = int64(len(characterSet))

	for i := 0; i < pwlength; i++ {
		max := big.NewInt(maxChar)
		bigN, err := crand.Int(crand.Reader, max)
		if err != nil {
			fmt.Println("crand.Int err", err)
		}
		n := bigN.Int64()
		identifier.WriteString(string(characterSet[n]))
	}
	return identifier.String()
}

func (h RequestHandler) IndexRoute(res http.ResponseWriter, req *http.Request) {
	// handle 404
	// if req.URL.Path != "/" {
	// 	h.ErrorRoute(res, req, http.StatusNotFound)
	// 	return
	// }

	// TODO figure out how tf to basic X)
	// res.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	if req.Method == "POST" {
		answer := myform.FormAnswer{}
		answer.ParsePost(req)
		fmt.Println("received a POST")
		var b []byte
		b, err := json.MarshalIndent(answer, "", "  ")
		if err != nil {
			fmt.Println("marshal err", err)
		} else {
			id := generateResponseIdentifier()
			responses[id] = string(b)
			persistData()
			// redirect to response page
			slug := fmt.Sprintf("/responder/%s", id)
			http.Redirect(res, req, slug, http.StatusFound)
		}
	} else if req.Method == "GET" {
		fmt.Println("GET")
		fmt.Fprintf(res, htmlContents)
	}
}

type ResponderData struct {
	Data string
}

const responseTemplate = `<!DOCTYPE html>
<html>
    <head>
    <title>Form submitted</title>
    <body>
    <h1>Response successful</h1>
    <p>Your response: </p>
    <pre>
    <code>
{{ .Data }}
    </code>
    </pre>
    <p><b>Bookmark this page</b> as a receipt or if you want to review what you responded some time in the future</p>
</body>
</html>`

const dataName = "latest-form-data.json"
func persistData() {
	b, err := json.Marshal(responses)
	if err != nil {
		fmt.Println("failure persisting data", err)
		return
	}
	err = os.WriteFile(dataName, b, 0777)
	if err != nil {
		fmt.Println("error writing persisted form data", err)
	}
}

func readPersistedData() {
	data, err := os.ReadFile(dataName)
	if errors.Is(err, os.ErrNotExist) {
		// no data yet probably, it's fine let's just return
		return
	}
	if err != nil {
		fmt.Println("error reading persisted form data", err)
		return
	}
	err = json.Unmarshal(data, &responses)
	if err != nil {
		fmt.Println("error unmarshalling persisted form data", err)
		return
	}
}

func Serve() {
	port := 7272
	handler := RequestHandler{}
	responses = make(map[string]string)
	readPersistedData()

	http.HandleFunc("/responder/", func(res http.ResponseWriter, req *http.Request) {
		id := strings.TrimPrefix(req.URL.Path, "/responder/")
			if val, ok := responses[id]; ok {
				t := template.Must(template.New("").Parse(responseTemplate))
				err := t.Execute(res, ResponderData{val})
				if errors.Is(err, syscall.EPIPE) {
					fmt.Println("recovering from broken pipe")
					return
				} else if err != nil {
					fmt.Println("err rendering reponder view", err)
				}
			} else {
				fmt.Fprintf(res, "No such form responder id")
			}
	})
	http.HandleFunc("/", handler.IndexRoute)

	// fileserver := http.FileServer(http.Dir("html/assets/"))
	// s.ServeMux.Handle("/assets/", http.StripPrefix("/assets/", fileserver))

	portstr := fmt.Sprintf(":%d", port)
	fmt.Println("Listening on port: ", portstr)
	http.ListenAndServe(portstr, nil)
}

func main () {
	Serve()
}
