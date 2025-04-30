all: install

ssh-attack:
	@cd src/attacks/ssh-bf; \
	echo "Compiling brute-force SSH attack binary..."; \
	go build -o ../../../bin/ssh-bf ssh.go; \
	chmod +x ../../../bin/ssh-bf; \
	echo "Brute-force SSH attack binary compiled successfully."; \

install:
	bash install.sh

.PHONY: all