# Mould

Mould is an all-inclusive form builder and server that uses a custom syntax for succinctly  declaring a form. 

Mould parses the incoming syntax and generates a form and a server that will serve the form and
receive its responses. Responses are saved in a json file and persisted between restarts of the
form server.

## Usage
```
go run main.go --input example-form-format.txt
go build server.go
./server
# Listening on port:  :7272
# Visit localhost:7272 in your browser to see the form in action! :)
```

## Flags

Generating the form page has a few options you can provide, other than the form input, such as
providing custom html and styles:

```
go run main.go --help

  -html-footer string
        a single html file containing all of the html that will be presented immediately below the form contents
  -html-header string
        a single html file containing all of the html that will be presented immediately above the form contents
  -input string
        a file containing the form format to generate a form server using
  -stylesheet string
        a single css file containing styles that will be applied to the form (fully replaces mould's default styling)
```

Change the port the server will run on by passing the `--port` flag:

```
go run server.go --help

  -port int
        the port to serve the form server on (default 7272)
``` 

## Example
```
form-title          = Nonsensical Form
form-desc           = Hey! Hello! This is a nonsensical form served by <untitled form>
form-bg             = wheat
form-titlecolor     = purple
form-fg             = black
input[Name]         = Preferred moniker
textarea[Address]   = Your fediverse residence, else null
number[Moni]#amount = min=1, max=100, value=1
radio[Sky type]     = Sunny, Rainy, Moony
```

* On the very left of the equals (=) sign is the **element**. Elements are a mix of html form elements (`input`, `textarea`) and elements for controlling themes (`form-bg`) or page titles (`form-title`)of the form (this latter group has the prefix `form-`). 
  * Examples: `input`, `radio`, `form-title`, `textarea`
* `= <stuff on the right side>` contains the **content** of the specified element. Typically, this will be used as
  part of the form element's placeholder, but in some cases (range, radio) it will set options,
  and in others (form-bg/form-fg) it will set colours or the page title (`form-title`).
* `[title]` sets the **title** that will be used for that form element's label
* `#key` sets an explicit **key**, which will be used instead of the title for things like keys on the input (useful if you want shorter html ids)

Currently supported html form elements:

* input[text] as `input`
* textarea as `textarea`
* input[range] as `range`
* input[number] as `number`
* radio buttons as `radio`
* input[hidden] as `hidden`
* required elements by prefixing a form element with `!`
    * example: `!input[Your favourite tea] = compulsory tea information here` 
* input[email] as `email`
    * the right-hand side of the email element is the regex pattern that validates it
    * `email[Email address] = .*@.*\..*
* paragraph elements as `form-paragraph`
* ~~checkboxes~~

## Basic auth: Password protection

Mould has support for [http basic
authentication](https://en.wikipedia.org/wiki/Basic_access_authentication) with the
`form-password` and `form-user` form options. To enable basic auth set at least
`form-password` in the form syntax input (default user: `mouldy`).

Basic auth should be used in combination with https / TLS secured connections to prevent
snooping the set password (http specifies that basic credentials are passed in plaintext with
the request).

## How does it work?
Messily! 

From the input syntax I generate the html page that will be served for the form as well as go
code, representing the response model. The generated go code is used to parse responses that
the form server receives.

All responses are saved in a local json file every time they come through. Respondents, on
submitting, are redirected to a static url containing their responses, should they forget
what they responded and want to refresh their memory.

The local json file is used to repopulate the form database between server restarts.

## Why did you do this?
Yes, why indeed

Well it's done now!!!
