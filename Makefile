INSTALL_DIR ?= ./out
BINARY = pkt-geo-mon

all:
	make clean
	mkdir -p ${INSTALL_DIR}
	go -C ./cmd build -o ../${INSTALL_DIR}/${BINARY}

clean:
	rm -f ${INSTALL_DIR}/${BINARY}
