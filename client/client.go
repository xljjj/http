package client

import (
	"HTTP/model"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Run Start the HTTP client
func Run(ip string, port int) {
	// Persistent connection
	addr := ip + ":" + strconv.Itoa(port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Cannot create a client")
		return
	} else {
		fmt.Println("Establish a connection. Your ip:", conn.LocalAddr().String())
	}
	URL := ""
	reader := bufio.NewReader(os.Stdin)
	for {
		// Since the correct connection has been established, there is no need to type the entire url
		fmt.Println("Type the path(exit to close):")
		line, _, _ := reader.ReadLine()
		URL = string(line)
		if URL == "exit" {
			break
		}
		// Though the input can be an entire url, check will be conducted
		// Parse the URL
		parsedURL, err := url.Parse(URL)
		if err != nil {
			fmt.Println("Invalid request")
			return
		}
		// Check the scheme,host and port
		schemeInput := parsedURL.Scheme
		hostInput := parsedURL.Hostname()
		portInput := parsedURL.Port()
		pathInput := parsedURL.Path
		if schemeInput != "" || hostInput != "" || portInput != "" {
			if (schemeInput != "http" && schemeInput != "https") || (hostInput != "localhost" && hostInput != "127.0.0.1") ||
				portInput != strconv.Itoa(port) {
				fmt.Println("Not the correct connection")
				continue
			}
		}
		// If the path does not start with /, add one
		if pathInput == "" || pathInput[0] != '/' {
			pathInput = "/" + pathInput
		}
		res, err := sendSimpleRequest(conn, pathInput)
		if err != nil {
			fmt.Println("Connection error")
		} else if res.Status/100 == 4 || res.Status/100 == 5 {
			fmt.Println(strconv.Itoa(res.Status) + " " + res.Desc)
		} else {
			switch res.ContentType {
			case "text/html":
				saveFile(res.Body, "index.html")
				ParseHTML(conn, res.Body)
			case "text/css":
				saveFile(res.Body, pathInput)
			case "image/jpeg", "image/x-icon":
				//showImage(res.Body)
				saveFile(res.Body, pathInput)
			default:
				fmt.Println("Cannot parse the type of data")
			}
		}
	}
}

// The request is only one line
func sendSimpleRequest(conn net.Conn, path string) (model.ResponseMessage, error) {
	// For convenience, only send the first line  Example: GET /index.html HTTP/1.1
	req := "GET " + path + " HTTP/1.1"
	_, err := conn.Write([]byte(req))
	if err != nil {
		return model.ResponseMessage{}, err
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return model.ResponseMessage{}, err
	} else {
		str := string(buf[:n])
		// The first packet can tell the content length
		res := model.ResponseMessage{}
		// Split the string into 6 parts(note there is an empty line)
		lines := strings.SplitN(str, "\n", 6)
		first := strings.SplitN(lines[0], " ", 3) //split the first line
		res.Version = first[0][5:]
		res.Status, _ = strconv.Atoi(first[1])
		res.Desc = first[2]
		res.ContentType = lines[1][14:]
		res.ContentLength, _ = strconv.Atoi(lines[2][16:])
		res.LastModified = lines[3][15:]
		res.Body = []byte(lines[5])
		// Receive the remaining data
		for len(res.Body) < res.ContentLength {
			n, err = conn.Read(buf)
			if err != nil {
				return model.ResponseMessage{}, err
			}
			res.Body = append(res.Body, buf[:n]...)
		}
		return res, err
	}
}

// ParseHTML using github.com/PuerkitoBio/goquery
func ParseHTML(conn net.Conn, data []byte) {
	reader := bytes.NewReader(data)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		fmt.Println(err)
		return
	}

	// I cannot find a library that has a better performance than the browser to show the resource
	// So only demonstrate the process of sending request and save the result locally

	// Request the .ico by default
	fmt.Println("Request:", "favicon.ico")
	res, err := sendSimpleRequest(conn, "favicon.ico")
	if err != nil {
		fmt.Println("Connection error")
	} else {
		handleDeterminedResponse(res, "favicon.ico")
	}

	// Request the .css
	doc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
		// Get href
		cssPath, exists := s.Attr("href")
		if exists {
			fmt.Println("Request:", cssPath)
			res, err := sendSimpleRequest(conn, cssPath)
			if err != nil {
				fmt.Println("Connection error")
			} else {
				handleDeterminedResponse(res, cssPath)
			}
		}
	})

	// Request the .jpg
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		// Get src
		src, exists := s.Attr("src")
		if exists {
			fmt.Println("Request:", src)
			res, err := sendSimpleRequest(conn, src)
			if err != nil {
				fmt.Println("Connection error")
			} else {
				handleDeterminedResponse(res, src)
			}
		}
	})
}

func handleDeterminedResponse(res model.ResponseMessage, path string) {
	if res.Status/100 == 4 || res.Status/100 == 5 {
		fmt.Println(strconv.Itoa(res.Status) + " " + res.Desc)
	} else {
		saveFile(res.Body, path)
	}
}

// Save the file locally
func saveFile(data []byte, path string) {
	ext := filepath.Ext(path)
	name := filepath.Base(path)
	switch ext {
	case ".html", ".css":
		err := os.WriteFile(name, data, 0666)
		if err != nil {
			fmt.Println("Cannot save the file")
		} else {
			fmt.Println("The file has been saved")
		}
	case ".jpg":
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			fmt.Println("Cannot decode the image")
		}

		file, err := os.Create(name)
		if err != nil {
			fmt.Println("Cannot create a file")
		}
		defer file.Close()

		if err := jpeg.Encode(file, img, nil); err != nil {
			fmt.Println("Cannot save the image")
		} else {
			fmt.Println("The image has been saved")
		}
	case ".ico":
		file, err := os.Create(name)
		if err != nil {
			fmt.Println("Cannot create a file")
		}
		defer file.Close()

		if _, err := file.Write(data); err != nil {
			fmt.Println("Cannot save the image")
		} else {
			fmt.Println("The image has been saved")
		}
	}
}

// -----------------------------------------------------------------------
// Show the jpg using github.com/hajimehoshi/ebiten/v2
// There is a problem in ebiten
// In one program, once you close a window, you can not open a new window
func showImage(data []byte) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		fmt.Println("Cannot parse the image")
		return
	}
	game := &Game{
		img: img,
	}
	ebiten.SetWindowTitle("Show the image")
	if err := ebiten.RunGame(game); err != nil {
		fmt.Println(err)
	}
}

type Game struct {
	img image.Image
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return errors.New("quit")
	}
	return nil
}
func (g *Game) Draw(screen *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	screen.DrawImage(ebiten.NewImageFromImage(g.img), opts)
}
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.img.Bounds().Dx(), g.img.Bounds().Dy()
}

//-----------------------------------------------------------------------
