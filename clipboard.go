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
}

func New() *Clipboard {
	c := new(Clipboard)
	return c
}

func (c *Clipboard) SetLocalPath(path string) {
	c.localPath = path
}

func (c *Clipboard) SetCloudPath(url, key, secret string) {
	c.url = url
	c.key = key
	c.secret = secret
}

func (c *Clipboard) ReadFrom(location string) string {
	if location == "local" {
		return readFile(c.localPath)
	}
	return ""
}

func (c *Clipboard) WriteTo(text *string, location string) {
	if location == "local" {
		writeFile(c.localPath, text)
	}
}

func (c *Clipboard) AppendTo(text *string, location string) {
	if location == "local" {
		appendFile(c.localPath, text)
	}
}

func (c *Clipboard) writeTo(text *string, location string) {
}

func (c *Clipboard) appendTo(text *string, location string) {

}

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
