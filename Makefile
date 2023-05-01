ROOT_DIR=/opt/zuccacm/
CONFIG_DIR=/etc/zuccacm/

all: zuccacm-server

# force to rebuild
force:

# build zuccacm-server
zuccacm-server: force
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/zuccacm-server

# install zuccacm-server
install:
	mkdir -p ${ROOT_DIR}
	install -m 0755 bin/zuccacm-server ${ROOT_DIR}
	install -m 0644 zuccacm-server.service ${ROOT_DIR}
	mkdir -p ${CONFIG_DIR}
	install -m 0644 zuccacm-server.yaml ${CONFIG_DIR}

# TODO test