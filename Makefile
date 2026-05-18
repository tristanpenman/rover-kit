DEST?=bin
GOARCH?=arm
GOARM?=6
TARGET?=stm32f4disco
FLAGS=-ldflags "-w"

ENV=env GOOS=linux GOARCH=$(GOARCH) GOARM=$(GOARM)

.PHONY: all camera-reader clean motor-control sonar-reader test tinygo-hello tinygo-sonar web-bridge

all: motor-control sonar-reader camera-reader tinygo-sonar tinygo-hello web-bridge

motor-control:
	$(ENV) go build $(FLAGS) -o $(DEST)/motor-control cmd/motor-control/main.go

sonar-reader:
	$(ENV) go build $(FLAGS) -o $(DEST)/sonar-reader cmd/sonar-reader/main.go

camera-reader:
	$(ENV) go build $(FLAGS) -o $(DEST)/camera-reader cmd/camera-reader/main.go

tinygo-sonar:
	tinygo build -target=$(TARGET) -o $(DEST)/sonar-$(TARGET).bin ./firmware/sonar

tinygo-hello:
	tinygo build -target=$(TARGET) -o $(DEST)/hello-$(TARGET).bin ./firmware/hello

web-bridge:
	$(ENV) go build $(FLAGS) -o $(DEST)/web-bridge cmd/web-bridge/main.go
	@rm -rf $(DEST)/static
	cp -R cmd/web-bridge/static $(DEST)/static

test:
	go test ./...

clean:
	rm -rf ${DEST}/{motor-control,sonar-reader,camera-reader,sonar-*.bin,hello-*.bin,web-bridge,web}
