build: 
	go build -o bin/swechat ./chat/

run: build
	./bin/swechat
