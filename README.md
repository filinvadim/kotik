# Installation
```
    go get github.com/filinvadim/kotik
```

# Proper usage
```
    go run main.go -p 5 http://google.com http://ya.com http://amazon.com 
```

# Wrong usage
```
    go run main.go http://google.com http://ya.com http://amazon.com -p 5 
```
Parallel flag won't be parsed and default will be used.

# Testing
```
    go test -count=1 -v ./... 
```
