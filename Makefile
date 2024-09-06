SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -o -name '*.sql')

all: dgen.exe

dgen.exe: $(SOURCES)
	go build -o $@ main.go
