test-container: docker run_container-structure-test run_dgoss

docker:
	docker build -t neumayer/dbwebapp .

container-structure-test:
	wget https://storage.googleapis.com/container-structure-test/latest/container-structure-test 
	chmod +x container-structure-test

run_container-structure-test: container-structure-test
	./container-structure-test test --image neumayer/dbwebapp --config test/dbwebapp.yaml 

dgoss:
	wget https://github.com/aelsabbahy/goss/releases/download/v0.3.5/dgoss
	chmod +x dgoss
run_dgoss: dgoss
	cd test; ../dgoss run -e DBUSER=u -e DBPASS=y  neumayer/dbwebapp

build:
	CGO_ENABLED=0 go build -o dbwebapp main.go

release:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o release/dbwebapp-linux-amd64-static
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o release/dbwebapp-darwin-amd64-static

clean:
	if [ -f dbwebapp ] ; then rm dbwebapp ; fi
	if [ -d release ] ; then rm -rf release ; fi
	if [ -f container-structure-test ] ; then rm container-structure-test ; fi
	if [ -f dgoss ] ; then rm dgoss ; fi
	docker rmi neumayer/dbwebapp

.PHONY: clean

