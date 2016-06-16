SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')


bin=bin/toc

all: $(bin)

$(bin): $(SOURCES)
	go build -o ${bin} .

.PHONY: clean
clean:
	rm -fv bin/*
