# Corpus

Very early prototype of a tool to work with a corpus of go code like [rsc's](https://github.com/rsc/corpus).

In its current form, it runs a command for every path in the corpus that looks like it has buildable go code in it. The results are stored in a sqlite3 database for analysis.

The command is given the current environment, replacing GOPATH with that specified with `-path` (defaults to current directory). This means you do not have to change your GOPATH to use go tools (e.g., `go vet`) in the corpus.

Results include the import path, error message (if any), and combined output of the command.

## Usage

```
Usage: corpus [flags] <command>
  -o string
        filename of sqlite3 db to store results it; it will be deleted before results are added (default "results.db")
  -path string
        root of corpus (similar to GOPATH) (default ".")
```

## Example

``` sh
$ wget https://storage.googleapis.com/go-corpus/go-corpus-0.01.tar.gz
$ tar -xvzf go-corpus-0.01.tar.gz
$ cd go-corpus-0.01
$ corpus -o vet.db go vet
```

The results are in `vet.db` and can be accessed using the `sqlite3` tool.