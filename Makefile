.PHONY: build

build: bootpowershell

bootpowershell: main.go
	go build .

clean:
	rm -f bootpowershell
