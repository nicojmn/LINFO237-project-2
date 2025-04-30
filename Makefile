all: install

ssh-attack:
	@cd src/attacks/ssh-bf; \
	echo "Compiling brute-force SSH attack binary..."; \
	go build -o ssh-bf.o ssh.go; \
	chmod +x ssh-bf.o; \
	echo "Brute-force SSH attack binary compiled successfully."; \

install:
	bash install.sh

.PHONY: all