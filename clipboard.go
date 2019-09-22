// Package clipboard read/write on clipboard
package clipboard

import (
	"io"
	"io/ioutil"
	"os"
)

type Clipboard struct {
	localPath string
	url       string
	key       string
	secret    string
	disabled  bool
}

func New() *Clipboard {
	c := new(Clipboard)
	return c
}

func (c *Clipboard) SetLocalPath(path string) {
	c.localPath = path
}

func (c *Clipboard) SetCloudPath(url, key, pass, secret string) {
	if key == "" || pass == "" {
		c.disabled = true
	}
	c.url = url
	c.key = key
	c.secret = secret
}

func (c *Clipboard) ReadFrom(location string) string {
	if location == "local" {
		return readFile(c.localPath)
	} else if location == "cloud" {
		return c.readCloud()
	}
	return ""
}

func (c *Clipboard) WriteTo(text *string, location string) string {
	msg := ""
	if location == "local" {
		writeFile(c.localPath, text)
	} else if location == "cloud" {
		msg = c.writeCloud(text)
	}
	return msg
}

func (c *Clipboard) AppendTo(text *string, location string) {
	if location == "local" {
		appendFile(c.localPath, text)
	}
}

// -----------------------------------
// Local clipboard
// -----------------------------------

func readFile(path string) string {
	var clip string

	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	clip = string(b)
	return clip
}

func writeFile(path string, text *string) {
	file, err := os.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.WriteString(file, *text)
	if err != nil {
		return
	}
	file.Sync()
}

func appendFile(path string, text *string) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return
	}
	_, err = io.WriteString(file, *text)
	if err != nil {
		return
	}
	file.Sync()
}

// -----------------------------------
// Cloud clipboard
// -----------------------------------

// Request and Response are encripted with user password
// Document is encript with user passphrase

// Cmd		: saveclip, getclip, uploadfile, downloadfile, uploadsettings, downloadsetting, uploadmicroide, downloadmicroide
// Document	: BASE64 encoded clip, file, settings or microide

type request struct {
	Cmd      string `json:"cmd"`
	Document string `json:"document,omitempty"`
}

// Status	: true=success, false==error
// ErrMsg	: Error message
// Document	: BASE64 encoded clip, file, settings or microide

type response struct {
	Status   bool   `json:"status"`
	ErrMsg   string `json:"errmsg"`
	Document string `json:"document"`
}

func (c *Clipboard) readCloud() string {
	if c.disabled {
		return ""
	}
	return ""
}

func (c *Clipboard) writeCloud(text *string) string {
	if c.disabled {
		return "Service not available"
	}
	return ""
}
