PROGRAM := video_recorder
DEST_DIR := /usr/local/bin

build:
	go build -o $(PROGRAM)

install:
	cp -p $(PROGRAM) $(DEST_DIR)

all: build install

clean:
	rm -f $(PROGRAM)

uninstall:
	rm -f $(DEST_DIR)/$(PROGRAM)
