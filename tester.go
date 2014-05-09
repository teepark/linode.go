package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"github.com/teepark/linode.go/linode"
)

func getAPIKey(path string) (string, error) {
	if path == "-" {
		buf := []byte{}
		if n, err := os.Stdin.Read(buf); err != nil {
			return "", err
		} else {
			return string(buf[:n]), nil
		}
	}

	var err error
	var buf []byte
	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if buf, err = ioutil.ReadFile(path); err != nil {
		return "", err
	} else {
		return string(buf), nil
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: linodeapi <api_key>")
		os.Exit(1)
	}

	apiKey, err := getAPIKey(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	responses, err := linode.Batch([]linode.APIRequest{
		linode.NewAPIRequest("test.echo", apiKey, map[string]interface{}{"foo": "bar"}),
		linode.NewAPIRequest("test.echo", apiKey, map[string]interface{}{"a": "b"}),
	})
	if err != nil {
		fmt.Println(err)
	} else {
		for _, resp := range responses {
			fmt.Println(resp.Data)
		}
	}
}
