PROGRAM := video_recorder
DEST_DIR := /usr/local/bin
IMG_DIR := /usr/local/share/images
IMG     := missed.jpg
build:
	go build -o $(PROGRAM)

install:
	cp -p $(PROGRAM) $(DEST_DIR)
	mkdir -p $(IMG_DIR)
	cp -p $(IMG) $(IMG_DIR)

all: build install

clean:
	rm -f $(PROGRAM)

uninstall:
	rm -f $(DEST_DIR)/$(PROGRAM)
