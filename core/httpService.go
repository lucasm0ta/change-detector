package core

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

type HttpService struct {
	Client http.Client
}

func NewHttpService() *HttpService {
	httpService := new(HttpService)
	httpService.Client = http.Client{}
	return httpService
}

func Body(doc *html.Node) (*html.Node, error) {
	var crawler func(*html.Node) *html.Node
	crawler = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode && node.Data == "body" {
			return node
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if result := crawler(child); result != nil {
				return result
			}
		}
		return nil
	}

	if body := crawler(doc); body != nil {
		return body, nil
	}
	return nil, errors.New("Missing <body> in the node tree")
}

func (httpService HttpService) GetBody(URL string) (*html.Node, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	htmlTree, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyNode, err := Body(htmlTree)
	if err != nil {
		return nil, err
	}
	return bodyNode, nil
}

func (httpService HttpService) GetNodeHash(node *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, node)
	hash := md5.New()
	io.WriteString(hash, buf.String())
	return fmt.Sprintf("%x", hash.Sum(nil))
}
