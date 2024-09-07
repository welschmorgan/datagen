TARGET_NAME := dgen
DIST_DIR := dist

TARGET := $(DIST_DIR)/$(TARGET_NAME)
SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -o -name '*.sql')

all: $(DIST_DIR)

$(DIST_DIR): $(TARGET)

$(TARGET): $(SOURCES)
	go build -o $@ main.go
