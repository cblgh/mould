package main

import (
	"fmt"
	"strings"
	"regexp"
	"path/filepath"
	"bufio"
	. "github.com/dave/jennifer/jen"
	"os"
)

/*
type FormContent struct {
	Title string
	Password string
	HeaderImage string
	Description string
}

type FormAnswer struct {
	Name string `json:"name"`
	Address string `json:"address"`
	StickerSheetAmount int `json:"sticker-sheet-amount"
	AccessToken string `json:"access-token"`
}
*/

func jsonTag (value string) map[string]string {
	return map[string]string{"json":value}
}

type genValue struct {
	element string
	title string
	value string
	key string
	options map[string]string
}

func parseFormat(format string) []genValue {
	pattern := regexp.MustCompile(`(form-\w+)|(\S*)(\[.*\])([#]\S+)?`)
	scanner := bufio.NewScanner(strings.NewReader(format))
	var genList []genValue
	for scanner.Scan() {
		line := scanner.Text()
		splitterIndex := strings.Index(line, "=")
		left := strings.TrimSpace(line[0:splitterIndex])

		var v genValue 
		v.value = strings.TrimSpace(line[splitterIndex+1:])
		matches := pattern.FindStringSubmatch(left)
		if len(matches) > 1 {
			v.element = strings.TrimSpace(matches[1])
		}
		if len(matches) > 2 && v.element == "" {
			v.element = strings.TrimSpace(matches[2])
		}
		if len(matches) > 3 && matches[3] != "" {
			// get everything except [thing] brackets
			v.title = matches[3][1:len(matches[3])-1]
		}
		if len(matches) > 4 && matches[4] != "" {
			// remove initial #
			v.key = strings.TrimSpace(matches[4][1:])
		}
		genList = append(genList, v)
	}
	return genList
}

const htmlTemplate = `<!DOCTYPE html>
<html>
	<head>
		<title>%s</title>
		<style>
		* {
			padding: 0;
			margin-bottom: 0.5rem;
		}
		div {
			display: grid;
			max-width: 600px;
			align-items: center;
		}
		</style>
	</head>
	<body>
		%s
	</body>
</html>`

func formatKeyAndTitle(v genValue) (string, string) {
	key := strings.ToLower(v.title)
	title := strings.ReplaceAll(strings.Title(v.title), " ", "")
	if len(v.key) > 0 {
		key = v.key
		title = strings.ReplaceAll(strings.Title(strings.ReplaceAll(v.key, "-", " ")), " ", "")
	}
	return key, title
}

const formPackageName = "myform"
func main() {
	var template []string
	var setPassword string
	var pageTitle string

	format := `form-title  = Merveilles Stickers
	form-desc   = Hey mervs, welcome to this form! Fill in the inputs below and then press submit!
	form-image  = ./header-image.png
	form-password   = hihi-stickertown
	input[Name]     = First and last name
	textarea[Address] = your postal address
	number[Sticker sheet amount]#amount = min=1, max=5, value=1
	input[The rabbit boat but backwards]#access-token = you know it.
	radio[Size]=Small, Medium, Large`

	values := parseFormat(format)


	f := NewFile(formPackageName)
	var contentBits []Code
	var answer []Code
	for _, input := range values {
		switch input.element {
		case "form-title":
			contentBits = append(contentBits, Id("Title").String())
			template = append(template, fmt.Sprintf(`<h1>%s</h1>`, input.value))
			pageTitle = input.value
		case "form-desc":
			contentBits = append(contentBits, Id("Description").String())
			template = append(template, fmt.Sprintf(`<p>%s</p>`, input.value))
		case "form-image":
			contentBits = append(contentBits, Id("Image").String())
			template = append(template, fmt.Sprintf(`<img src="%s">`, input.value))
		case "form-password":
			setPassword = input.value
			// information used for basic auth, limiting access to the form
			contentBits = append(contentBits, Id("Password").String())
		}
	}

	template = append(template, `<form action="/" method="post">`)
	for _, input := range values {
		switch input.element {
		case "textarea":
			key, title := formatKeyAndTitle(input)
			template = append(template, "<div>")
			template = append(template, fmt.Sprintf(`<label for="%s">%s</label>`, key, title))
			el := fmt.Sprintf(`<textarea placeholder="%s" name="%s"></textarea>`, input.value, key)
			template = append(template, el)
			template = append(template, "</div>")
			answer = append(answer, Id(title).String().Tag(jsonTag(key)))
		case "input":
			key, title := formatKeyAndTitle(input)
			template = append(template, "<div>")
			template = append(template, fmt.Sprintf(`<label for="%s">%s</label>`, key, input.title))
			el := fmt.Sprintf(`<input type="text" placeholder="%s" name="%s"/>`, input.value, key)
			template = append(template, el)
			template = append(template, "</div>")
			answer = append(answer, Id(title).String().Tag(jsonTag(key)))
		case "number":
			optionsList := strings.Split(input.value, ",")
			var options string
			template = append(template, "<div>")
			for _, optionPair := range optionsList {
				optionPair = strings.TrimSpace(optionPair)
				parts := strings.Split(optionPair, "=")
				options += fmt.Sprintf(`%s="%s" `,parts[0], parts[1])
			}
			key, title := formatKeyAndTitle(input)
			template = append(template, fmt.Sprintf(`<label for="%s">%s</label>`, key, title))
			el := fmt.Sprintf(`<input type="number" %s name="%s"/>`, options, key)
			template = append(template, el)
			template = append(template, "</div>")
			answer = append(answer, Id(title).String().Tag(jsonTag(key)))
		case "range":
			optionsList := strings.Split(input.value, ",")
			var options string
			template = append(template, "<div>")
			for _, optionPair := range optionsList {
				optionPair = strings.TrimSpace(optionPair)
				parts := strings.Split(optionPair, "=")
				options += fmt.Sprintf(`%s="%s" `,parts[0], parts[1])
			}
			key, title := formatKeyAndTitle(input)
			template = append(template, fmt.Sprintf(`<label for="%s">%s</label>`, key, title))
			el := fmt.Sprintf(`<input type="range" %s name="%s"/>`, options, key)
			template = append(template, el)
			template = append(template, "</div>")
			answer = append(answer, Id(title).String().Tag(jsonTag(key)))
		case "radio":
			options := strings.Split(input.value, ",")
			key, title := formatKeyAndTitle(input)

			template = append(template, "<div>")
			template = append(template, fmt.Sprintf(`<span>%s</span>`, title))
			for i, val := range options {
				options[i] = strings.TrimSpace(val)
				radioValue := strings.ToLower(options[i])
				radioId := fmt.Sprintf(`%s-option-%s`, key, radioValue)
				template = append(template, "<span>")
				template = append(template, fmt.Sprintf(`<label for="%s">%s</label>`, radioId, options[i]))
				el := fmt.Sprintf(`<input type="radio" id="%s" value="%s" name="%s"/>`, radioId, radioValue, key)
				template = append(template, el)
				template = append(template, "</span>")

			}
			template = append(template, "</div>")
			answer = append(answer, Id(title).String().Tag(jsonTag(key)))
		}
	}

	template = append(template, `<div><button type="submit">Submit</button></div>`)
	template = append(template, "</form>")

	f.Const().Id("BasicPassword").Op("=").Lit(setPassword)
	f.Type().Id("FormContent").Struct(contentBits...)
	f.Type().Id("FormAnswer").Struct(answer...)

	err := os.MkdirAll(formPackageName, 0777)
	if err != nil {
		fmt.Println("err mkdirall", err)
	}
	generatedCode := fmt.Sprintf("%#v", f)
	genCodeErr := os.WriteFile(filepath.Join(formPackageName, "generated-form-model.go"), []byte(generatedCode), 0777)
	if genCodeErr != nil {
		fmt.Println(genCodeErr)
	}
	htmlContents := fmt.Sprintf(htmlTemplate, pageTitle, strings.Join(template, "\n\t\t"))
	indexWriteErr := os.WriteFile("index-template.html", []byte(htmlContents), 0777)
	if indexWriteErr != nil {
		fmt.Println(indexWriteErr)
	}
}
