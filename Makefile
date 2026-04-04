FLAGS=-ldflags "-w"

ENV=env GOOS=linux GOARCH=arm GOARM=6

.PHONY: all clean bin/motor-control bin/sonar-reader bin/web-bridge

all: bin/motor-control bin/sonar-reader bin/web-bridge

bin/motor-control: cmd/motor-control/main.go
	$(ENV) go build $(FLAGS) -o bin/motor-control cmd/motor-control/main.go

bin/sonar-reader: cmd/sonar-reader/main.go
	$(ENV) go build $(FLAGS) -o bin/sonar-reader cmd/sonar-reader/main.go

bin/web-bridge: cmd/web-bridge/main.go
	$(ENV) go build $(FLAGS) -o bin/web-bridge cmd/web-bridge/main.go

clean:
	rm -f bin/motor-control bin/sonar-reader bin/web-bridge
