// +build ignore

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var ZIP *zip.Writer

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func build() {
	cmd := exec.Command("go", "build", "-o", filepath.Join(".bin", "run"), ".")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH=amd64")
	check(cmd.Run())

	AddDir(".bin")

	AddGlob("Docker*")
	AddGlob("LICENSE*")

	// add your own additional folders here
	AddDir("assets")
	AddDir("templates")
}

func main() {
	os.Mkdir(".bin", 0755)
	os.Mkdir(".deploy", 0755)

	filename := fmt.Sprintf("%s.zip", time.Now().Format("2006-01-02-15-04"))

	file, err := os.Create(filepath.Join(".deploy", filename))
	check(err)
	defer file.Close()

	fmt.Println("Creating:", filename)

	ZIP = zip.NewWriter(file)
	defer ZIP.Close()

	build()
}

// filename with forward slashes
func AddFile(filename string) {
	fmt.Printf("  %-40s", filename)
	defer fmt.Println("+")

	file, err := os.Open(filepath.FromSlash(filename))
	check(err)
	defer file.Close()

	w, err := ZIP.Create(filename)
	check(err)
	_, err = io.Copy(w, file)
	check(err)
}

// glob with forward slashes
func AddGlob(glob string) {
	fmt.Printf("G %v\n", glob)
	matches, err := filepath.Glob(filepath.FromSlash(glob))
	check(err)
	for _, match := range matches {
		AddFile(filepath.ToSlash(match))
	}
}

// dir with forward slashes
func AddDir(dir string) {
	fmt.Printf("D %v\n", dir)
	check(filepath.Walk(filepath.FromSlash(dir),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			AddFile(filepath.ToSlash(path))
			return nil
		}))
}
