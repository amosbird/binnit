GO=go

all: binnit

binnit: binnit.go config.go
	CGO_ENABLED=0 $(GO) build -o binnit binnit.go config.go
