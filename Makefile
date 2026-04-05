DEST?=bin
GOARCH?=arm
GOARM?=6

FLAGS=-ldflags "-w"

ENV=env GOOS=linux GOARCH=$(GOARCH) GOARM=$(GOARM)

.PHONY: all clean motor-control sonar-reader test web-bridge

all: motor-control sonar-reader web-bridge

motor-control:
	$(ENV) go build $(FLAGS) -o $(DEST)/motor-control cmd/motor-control/main.go

sonar-reader:
	$(ENV) go build $(FLAGS) -o $(DEST)/sonar-reader cmd/sonar-reader/main.go

web-bridge:
	$(ENV) go build $(FLAGS) -o $(DEST)/web-bridge cmd/web-bridge/main.go
	@rm -rf $(DEST)/web
	cp -R cmd/web-bridge/web $(DEST)/web

test:
	go test ./...

clean:
	rm -rf ${DEST}/{motor-control,sonar-reader,web-bridge,web}
