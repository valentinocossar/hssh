package models

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// Host ...
type host struct {
	path        string
	content     string
	connections []Connection
	channel     chan Connection
}

// IHost ...
type IHost interface {
	ReadFile()
	ParseRow(string) Connection
	Parse() []Connection
	GetPath() string
	GetContent() string
	GetConnectionsCount() int
	GetConnections() []Connection
	ProvideViaChannel(*chan Connection)
	Create([]byte) error
}

func (h *host) ReadFile() {
	contentBytes, err := ioutil.ReadFile(h.path)
	if err != nil {
		return
	}

	h.content = string(contentBytes)
}

func (h *host) Create(content []byte) error {
	filePathSplitted := strings.Split(h.path, "/")
	folderPath := filePathSplitted[0 : len(filePathSplitted)-1]
	err := os.MkdirAll(strings.Join(folderPath, "/"), os.ModePerm)

	file, err := os.Create(h.path)
	if err != nil {
		return err
	}

	defer file.Close()

	if _, err := file.Write(content); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	return nil
}

func (h *host) getAttributeFromRow(attribute string, row string) string {
	rgx := regexp.MustCompile(attribute + "(|\\s)")
	return strings.Trim(rgx.ReplaceAllString(row, ""), " ")
}

func (h *host) ParseRow(hostRaw string) Connection {
	connection := Connection{}
	for _, attribute := range strings.Split(hostRaw, "\n") {
		attribute = strings.Trim(attribute, " ")
		attribute = strings.ToLower(attribute)

		if attribute == "" {
			continue
		}

		if ok, err := regexp.MatchString(`^hostname `, attribute); err == nil && ok {
			connection.Hostname = h.getAttributeFromRow("hostname", attribute)
		}

		if ok, err := regexp.MatchString(`^user `, attribute); err == nil && ok {
			connection.User = h.getAttributeFromRow("user", attribute)
		}

		if ok, err := regexp.MatchString(`^port `, attribute); err == nil && ok {
			connection.Port = h.getAttributeFromRow("port", attribute)
		}

		if ok, err := regexp.MatchString(`^identityfile `, attribute); err == nil && ok {
			connection.IdentityFile = h.getAttributeFromRow("identityfile", attribute)
		}

		if ok, err := regexp.MatchString(`^host `, attribute); err == nil && ok {
			connection.Name = h.getAttributeFromRow("host", attribute)
		}
	}

	return connection
}

func (h *host) ProvideViaChannel(channel *chan Connection) {
	if channel == nil {
		return
	}

	for _, connection := range h.connections {
		go func(c Connection) {
			*channel <- c
		}(connection)
	}
}

func (h *host) Parse() []Connection {
	content := strings.TrimSpace(h.content)

	// Remove comments
	content = regexp.MustCompile("(?m)^(|\\s+)#.*").ReplaceAllString(content, "")

	// Remove empty lines
	content = regexp.MustCompile("[\t\r\n]+").ReplaceAllString(content, "\n")

	// Apply a marker for splitting logic
	content = regexp.MustCompile("Host ").ReplaceAllString(content, "!!Host ")

	// Split content into hosts
	hosts := strings.Split(content, "!!")

	for x, host := range hosts {
		if x == 0 {
			continue
		}

		connection := h.ParseRow(host)
		if !connection.IsAllowed() {
			continue
		}

		if !connection.IsWellConfigured() {
			continue
		}

		h.connections = append(h.connections, connection)
	}

	return h.connections
}

func (h *host) GetPath() string {
	return h.path
}

func (h *host) GetContent() string {
	return h.content
}

func (h *host) GetConnections() []Connection {
	return h.connections
}

func (h *host) GetConnectionsCount() int {
	return len(h.connections)
}

// NewHost ...
func NewHost(path string) IHost {
	return &host{
		path: path,
	}
}
