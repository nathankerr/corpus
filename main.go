package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.SetFlags(log.Lshortfile)

	path := flag.String("path", ".", "root of corpus (similar to GOPATH)")
	dbFilename := flag.String("o", "results.db", "filename of sqlite3 db to store results it; it will be deleted before results are added")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: corpus [flags] <command>")
		flag.PrintDefaults()
		return
	}

	command := flag.Args()

	// *path needs to be absolute because it becomes the GOPATH
	// of whatever command is ran
	var err error
	*path, err = filepath.Abs(*path)
	if err != nil {
		log.Fatalln(err)
	}

	// get the set of valid import paths in *path
	packages := map[string]struct{}{}
	src := filepath.Join(*path, "src")
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// I really want import paths that are "buildable"
		// this is just a proxy for that
		if filepath.Ext(path) == ".go" {
			rel, err := filepath.Rel(src, path)
			if err != nil {
				log.Fatalln(err)
			}

			packages[filepath.Dir(rel)] = struct{}{}
		}

		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}

	// create and setup results db

	// removing the file avoids needing to manage its contents
	// e.g., schema, etc.
	// this should really be dealt with better to enable continuing a run, etc
	err = os.Remove(*dbFilename)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalln(err)
		}
	}

	db, err := sql.Open("sqlite3", *dbFilename)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = db.Exec("create table results (path text, output blob, err text)")
	if err != nil {
		log.Fatalln(err)
	}

	// run the command on each package
	bar := pb.StartNew(len(packages))
	defer bar.Finish()
	for pkg := range packages {
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Dir = filepath.Join(src, pkg)

		// use the current environment, but replace GOPATH
		env := os.Environ()
		for i, str := range env {
			if strings.HasPrefix(str, "GOPATH") {
				env[i] = fmt.Sprintf("GOPATH=%v", *path)
			}
		}
		cmd.Env = env

		// store the results
		output, err := cmd.CombinedOutput()
		if err != nil {
			_, err = db.Exec("insert into results (path, output, err) values (?, ?, ?)", pkg, output, err.Error())
		} else {
			_, err = db.Exec("insert into results (path, output) values (?, ?)", pkg, output)
		}
		if err != nil {
			log.Fatalln(err)
		}

		bar.Increment()
	}
}
