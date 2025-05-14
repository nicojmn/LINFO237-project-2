all: install

ssh-attack:
	@cd src/attacks/ssh-bf; \
	echo "Compiling brute-force SSH attack binary..."; \
	go build -o ../../../bin/ssh-bf ssh.go; \
	chmod +x ../../../bin/ssh-bf; \
	echo "Brute-force SSH attack binary compiled successfully."; \


syn-flood:
	@cd src/attacks/syn-flood; \
	echo "Compiling SYN flood attack binary..."; \
	go build -o ../../../bin/syn-flood syn-flood.go; \
	chmod +x ../../../bin/syn-flood; \
	echo "SYN flood attack binary compiled successfully."; \


dos-attack:
	@cd src/attacks/rf-dos; \
	echo "Compiling Reflected DoS attack binary..."; \
	go build -o ../../../bin/rf-dos rf-dos.go; \
	chmod +x ../../../bin/rf-dos; \
	echo "Reflected DoS attack binary compiled successfully."; \

port-scan:
	@cd src/attacks/rf-dos; \
	echo "Compiling Port scanning attack binary..."; \
	go build -o ../../../bin/rf-dos port-scan.go; \
	chmod +x ../../../bin/rf-dos; \
	echo "Port scanning attack binary compiled successfully."; \


install:
	bash install.sh
	
attacks: install ssh-attack syn-flood dos-attack
	@echo "All attacks compiled successfully."

clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f bin.zip
	@rm -f project.zip
	@find . -name "*.o" -delete
	@echo "Cleaned up successfully."

# Tu aimes mon usine à gaz ? Elle consomme plus que l'Allemagne
zip: ssh-attack syn-flood
	@echo "Zipping project files..."
	@mkdir -p /tmp/project
	@rsync -a --exclude=bin/ --exclude='*.zip' --exclude='.git*' --exclude='*.pdf' --exclude='*.md' --exclude='.vscode' ./ /tmp/project/
	@cd /tmp && zip -r project.zip project
	@mv /tmp/project.zip ./project.zip
	@rm -rf /tmp/project /tmp/project.zip
	@echo "Project files zipped successfully."

upload: clean zip
	@echo "Uploading zip file to remote server..."
	scp project.zip mininet-vm:/home/mininet/
	@echo "Zip file uploaded successfully."
	

.PHONY: all clean ssh-attack
