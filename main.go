package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

const (
	gitignoreFilesDirName = ".got"
	gitignoreRepoURL      = "https://api.github.com/repos/github/gitignore/tarball"
)

func help() {
	fmt.Println("Usage: got generate <name1> <name2> ...")
}

func getIgnores() (map[string]string, error) {
	// Check if .got folder exists in the home holder.
	// If it doens't, download from github gitignore repository.
	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	ignoreFilesDir := path.Join(homeDir, gitignoreFilesDirName)
	_, err = os.Stat(ignoreFilesDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		// Create the directory.
		if err = os.Mkdir(ignoreFilesDir, 0700); err != nil {
			return nil, err
		}

		// Create the archive file.
		archiveFilename := path.Join(ignoreFilesDir, "archive.tar.gz")
		var archiveFile *os.File
		archiveFile, err = os.Create(archiveFilename)
		if err != nil {
			return nil, err
		}
		defer archiveFile.Close()

		// Download the archive file.
		var getResp *http.Response
		getResp, err = http.Get(gitignoreRepoURL)
		if err != nil {
			return nil, err
		}
		defer getResp.Body.Close()

		if _, err = io.Copy(archiveFile, getResp.Body); err != nil {
			return nil, err
		}
		archiveFile.Close()
		getResp.Body.Close()

		// Unzip the archive file.
		unzipCmd := exec.Command("tar", "xvf", archiveFilename, "-C", ignoreFilesDir)
		unzipCmd.Stdout = os.Stdout
		if err = unzipCmd.Run(); err != nil {
			return nil, err
		}
	}

	// Walk the folder to get a map from name to file path.
	nameToPath := make(map[string]string)
	if err = filepath.Walk(ignoreFilesDir, func(path string, info os.FileInfo, err error) error {
		filename := info.Name()
		if endIndex := strings.Index(filename, ".gitignore"); endIndex != -1 {
			name := strings.ToLower(filename[0:endIndex])
			nameToPath[name] = path
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return nameToPath, nil
}

func main() {
	// Validate arguments.
	args := os.Args
	if len(args) <= 2 || args[1] != "generate" {
		help()
		return
	}
	requests := make(map[string]struct{})
	for i := 2; i < len(args); i++ {
		requests[strings.ToLower(args[i])] = struct{}{}
	}

	// Get a map from names to paths.
	allIgnores, err := getIgnores()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Generate the .gitignore file.
	var ignoreContent string
	for request := range requests {
		ignoreFile, ok := allIgnores[request]
		if !ok {
			fmt.Printf("%s cannot be recoginized\n", request)
			return
		}
		var bytes []byte
		bytes, err = ioutil.ReadFile(ignoreFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		ignoreContent += string(bytes) + "\n"
	}

	if err = ioutil.WriteFile(".gitignore", []byte(ignoreContent), 0700); err != nil {
		fmt.Println(err)
		return
	}
}
