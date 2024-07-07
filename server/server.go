package server

import (
	"HTTP/model"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Run Start the HTTP server
func Run(port int) {
	fmt.Println("Server port:", port)
	prefix := getParentDirectory(getCurrentDirectory()) + "/web" // The prefix of the resources
	fmt.Println("Resource prefix:", prefix)
	addr := "127.0.0.1:" + strconv.Itoa(port)
	// Create a server
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Cannot create a server")
		return
	}
	defer l.Close()
	// Wait for connection
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("A new client ip address:", conn.RemoteAddr().String())
			go handleConnection(conn, prefix) // Create a goroutine to handle it
		}
	}
}

func handleConnection(conn net.Conn, prefix string) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("A client exits")
			break
		} else {
			fmt.Println(conn.RemoteAddr().String() + ":")
			fmt.Println(string(buf[:n]))
			handleRequest(conn, string(buf[:n]), prefix)
		}
	}
}

// Handle HTTP request
func handleRequest(conn net.Conn, req string, prefix string) {
	// Only use the first line currently. Example:GET /index.html HTTP/1.1
	lines := strings.Split(req, "\n")
	resource := strings.Split(lines[0], " ")[1]
	if resource == "/" || resource == "/index" {
		resource = "index.html"
	}
	res := createResponseMessage(prefix, resource)
	_, err := conn.Write([]byte(res.String()))
	if err != nil {
		fmt.Println("Connection error")
	}
}

func createResponseMessage(prefix string, path string) model.ResponseMessage {
	res := model.ResponseMessage{Version: "1.1"}
	// Get the extension of the file
	ext := filepath.Ext(path)
	// Only accept .html .css .jpg .ico
	switch ext {
	case ".html":
		res.ContentType = "text/html"
	case ".css":
		res.ContentType = "text/css"
	case ".jpg":
		res.ContentType = "image/jpeg"
	case ".ico":
		res.ContentType = "image/x-icon"
	default:
		res.Status = 404
		res.Desc = "Not Found"
		return res
	}
	// Get the resource
	data, err := os.ReadFile(prefix + "/" + path)
	if err != nil {
		res.Status = 404
		res.Desc = "Not Found"
		return res
	}
	res.Status = 200
	res.Desc = "OK"
	res.ContentLength = len(data)
	res.Body = data
	// Send last modified time(Do not want the browser to send request continuously)
	// Assume the resource do not change, not disable the cache in the response message
	// Example: Sat, 20 Apr 2024 08:00:00 GMT
	res.LastModified = getLastModifiedTime(prefix + "/" + path)
	return res
}

// Get current directory and parent directory  Reference: https://studygolang.com/articles/3421

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

func getParentDirectory(directory string) string {
	return substr(directory, 0, strings.LastIndex(directory, "/"))
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

// Get Last modified time of a file
func getLastModifiedTime(path string) string {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "Sat, 20 Apr 2024 08:00:00 GMT" // Use a default time for convenience
	} else {
		modTime := fileInfo.ModTime()
		formattedTime := modTime.Format("Mon, 02 Jan 2006 15:04:05 MST")
		return formattedTime
	}
}
