SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

dgen.exe: $(SOURCES)
	go build -o $@ main.go
