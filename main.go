package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {
	// Check arguments.
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: got <type>...")
		return
	}

	// Get all contents from gitignore.io
	var gitignores string
	for _, target := range args {
		gitignore, err := getGitIgnore(target)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		gitignores += gitignore
	}

	// Create a .gitignore file.
	gitIgnoreFile, err := os.Create(".gitignore")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	_, err = gitIgnoreFile.WriteString(gitignores)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

func getGitIgnore(target string) (string, error) {
	httpResponse, err := http.Get("https://www.gitignore.io/api/" + target)
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()
	bytesContent, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return "", err
	}
	content := string(bytesContent)
	if strings.Contains(content, "ERROR") {
		return "", fmt.Errorf("%s is undefined.", target)
	}
	return content, nil
}
