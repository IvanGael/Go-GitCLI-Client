Git CLI Client Built in Go

### Features
- git init
- git config
- git add
- git commit 
- git status
- git diff
- git logs


# Install

**Source**

1. Requirement

```
Go 1.21.4
```

2. Build binary

```
go build main.go;
```

3. Run the build

```
main
```


# Usage

Type from commandline

Example:

```
go run main.go init
go run main.go config AUTHOR_USERNAME AUTHOR_EMAIL
go run main.go add FILENAME
go run main.go commit MESSAGE
go run main.go status
go run main.go diff
go run main.go logs
```
