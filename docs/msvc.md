# MSVC
This page describes some details for working with the MSVC (Microsoft compiler) toolchain.

## Create import library with MSVC
If you want to create your import library (.lib) with MSVC tools you can use the following steps:

```shell
# build library
go mod tidy
go build -buildmode=c-shared -o ./cinterface/libcinterface.dll github.com/image357/password/cinterface
cd cinterface

# print exports
dumpbin /exports libcinterface.dll

# then: manually create libcinterface.def with EXPORTS heading

# create import library
lib /DEF:libcinterface.def /MACHINE:x64 /OUT:libcinterface.lib
```

Tested with POSIX-Threads and UCRT for MinGW-w64.
