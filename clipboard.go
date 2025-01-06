// Package clipboard read/write on clipboard
package clipboard

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Clipboard struct {
	localPath string
	url       string
	apikey    string
	apipass   string
	secret    string
	disabled  bool
}

// -----------------------------------
// Clipboard Setup
// -----------------------------------

func New() *Clipboard {
	c := new(Clipboard)
	return c
}

func (c *Clipboard) SetLocalPath(path string) {
	c.localPath = path
}

func (c *Clipboard) SetCloudPath(url, apikey, apipass, secret string) error {
	c.disabled = true
	if url == "" || apikey == "" || apipass == "" || secret == "" {
		return errors.New("incomplete cloud connection information")
	} else if !strings.Contains(url, "https") && !strings.Contains(url, "localhost") {
		return errors.New("incorrect url")
	}
	c.disabled = false
	c.url = url
	c.apikey = apikey
	c.apipass = apipass
	c.secret = secret
	return nil
}

// -----------------------------------
// Clipboard
// -----------------------------------

func (c *Clipboard) ReadFrom(location, cmd string) string {
	if location == "local" {
		return readFile(c.localPath)
	} else if location == "cloud" {
		return c.readCloud(cmd)
	}
	return ""
}

func (c *Clipboard) WriteTo(text *string, location, cmd string) string {
	msg := ""
	if location == "local" {
		writeFile(c.localPath, text)
	} else if location == "cloud" {
		msg = c.writeCloud(cmd, text)
	} else {
		return ""
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
	b, _ := io.ReadAll(file)
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

func (c *Clipboard) CloudDisabled() bool {
	return c.disabled
}

func (c *Clipboard) SetUpCloudService() string {
	empty := ""
	return c.writeCloud("setup", &empty)
}

func (c *Clipboard) ChangeCloudPassword(newpass string) string {
	if newpass == "" {
		return "Password can not be empty"
	}
	return c.writeCloud("chgpass", &newpass)
}

func (c *Clipboard) ResetCloudService() string {
	empty := ""
	return c.writeCloud("reset", &empty)
}

// Cmd		: clip, file, setting, miide, setup
// Document	: Encrypted, BASE64 clip, file, settings or miide

type request struct {
	Cmd      string `json:"cmd"`
	Key      string `json:"apikey"`
	Pass     string `json:"apipass"`
	NewPass  string `json:"newpass,omitempty"`
	Document string `json:"document,omitempty"`
}

// Success	: true, false
// ErrMsg	: Error message
// Document	: BASE64 encoded clip, file, settings or miide

type response struct {
	Success  bool   `json:"success"`
	ErrMsg   string `json:"errmsg"`
	Document string `json:"document"`
}

func (c *Clipboard) readCloud(cmd string) string {
	if c.disabled {
		return ""
	}
	// Create request and send to server
	var jresp response
	req := new(request)
	req.Cmd = cmd
	req.Key = c.apikey
	req.Pass = c.apipass
	jreq, _ := json.Marshal(req)
	hreq, _ := http.NewRequest("POST", c.url+"/get", bytes.NewBuffer(jreq))
	hreq.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Second * 20,
	}
	resp, err := client.Do(hreq)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	sresp, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	if err = json.Unmarshal(sresp, &jresp); err != nil {
		return ""
	}
	if !jresp.Success {
		return ""
	}
	text := DecryptData(c.secret, jresp.Document)
	return string(text)
}

func (c *Clipboard) writeCloud(cmd string, text *string) string {
	if c.disabled {
		return "wC:Service not available"
	}
	var jresp response
	req := new(request)
	req.Cmd = cmd
	req.Key = c.apikey
	req.Pass = c.apipass
	if cmd == "chgpass" {
		req.NewPass = *text
		*text = ""
	}
	if *text != "" {
		req.Document = EncryptData(c.secret, *text)
	}
	jreq, _ := json.Marshal(req)
	hreq, _ := http.NewRequest("POST", c.url+"/put", bytes.NewBuffer(jreq))
	hreq.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Second * 20,
	}
	resp, err := client.Do(hreq)
	if err != nil {
		return "wC1:" + err.Error()
	}
	defer resp.Body.Close()
	sresp, err := io.ReadAll(resp.Body)
	if err != nil {
		return "wC2:" + err.Error()
	}
	if err = json.Unmarshal(sresp, &jresp); err != nil {
		return "wC3:" + err.Error()
	}
	if !jresp.Success {
		return jresp.ErrMsg
	}
	if cmd == "chgpass" {
		c.apipass = req.NewPass
	}
	return ""
}

// Crypt routines

func EncryptData(keyString, data string) string {
	plaintext := []byte(data)
	key, _ := hex.DecodeString(keyString) // Convert the key to bytes
	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize] // Initialization vector
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return ""
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// Return the encoded hex string
	return hex.EncodeToString(ciphertext)
}

func DecryptData(keyString, ecdata string) string {
	key, _ := hex.DecodeString(keyString)
	ciphertextBytes, _ := hex.DecodeString(ecdata)

	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}

	if len(ciphertextBytes) < aes.BlockSize {
		return ""
	}

	iv := ciphertextBytes[:aes.BlockSize]
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertextBytes, ciphertextBytes)

	return string(ciphertextBytes)
}
