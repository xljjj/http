package model

import (
	"fmt"
)

// ResponseMessage HTTP response message  Can add more parameters
type ResponseMessage struct {
	Version       string
	Status        int
	Desc          string
	ContentType   string
	ContentLength int
	LastModified  string
	Body          []byte
}

// Convert the struct to string
func (res *ResponseMessage) String() string {
	return fmt.Sprintf("HTTP:%s %d %s\nContent-Type: %s\nContent-Length: %d\nLast-Modified: %s\n\n%s",
		res.Version, res.Status, res.Desc, res.ContentType, res.ContentLength, res.LastModified, res.Body)
}
