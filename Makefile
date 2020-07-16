build: bin
	go build -o bin/calabash cmd/calabash.go  
	
bin:
	mkdir -p bin