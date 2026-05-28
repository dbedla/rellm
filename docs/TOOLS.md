# Project tools usage examples

This document defines practical rules and conventions tools usage, with examples.


## make 
always use `make` to build the project.

- to build a project, regularly use 
``` bash
make go-build
```

- to make clean build
```bash
make go-clean-build
```

- to run tests
```bash
make go-test
```

To run specific target from cmd use proper make target
never call `go run` directly.