docker:
	docker build -t neumayer/dbwebapp .

build:
	CGO_ENABLED=0 go build -o dbwebapp main.go

release:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o release/dbwebapp-linux-amd64-static
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o release/dbwebapp-darwin-amd64-static

clean:
	if [ -f dbwebapp ] ; then rm dbwebapp ; fi
	if [ -d release ] ; then rm -rf release ; fi

.PHONY: clean
