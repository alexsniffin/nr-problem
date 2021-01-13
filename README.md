# nr-problem

**How to run your program:**

Tested and built with Go 1.15
```
# run one file
go run main.go -fpath origin-of-species.txt

# run multiple files
go run main.go -fpath origin-of-species.txt -fpath etc.txt

# stdin
cat origin-of-species.txt | go run main.go

# build binary
go build

# docker build
docker build . -t nr-problem

# docker stdin
cat origin-of-species.txt | docker run -i nr-problem cat
```

**What you would do next, given more time (if anything)?**

- Modularize the logic into a organized project structure
- Better test coverage and mock file system for better quality of unit tests
- Unicode support

**Are there bugs that you are aware of?**

- N/A