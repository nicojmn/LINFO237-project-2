all: install

ssh-attack:
	@cd src/attacks/ssh-bf; \
	echo "Compiling brute-force SSH attack binary..."; \
	go build -o ../../../bin/ssh-bf ssh.go; \
	chmod +x ../../../bin/ssh-bf; \
	echo "Brute-force SSH attack binary compiled successfully."; \

install:
	bash install.sh

clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f bin.zip
	@echo "Cleaned up successfully."

bin-zip: ssh-attack
	@echo "Creating zip file for binaries..."
	@cd bin/; \
	zip -r ../bin.zip *;
	@echo "Zip file created successfully."

.PHONY: all clean