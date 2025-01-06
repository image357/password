![Go workflow](https://github.com/image357/password/workflows/Go/badge.svg)
![CMake workflow](https://github.com/image357/password/workflows/CMake/badge.svg)

# simple-password-manager

Ever needed a simple key-value store with an encryption backend to handle your app passwords?

**Look no further!** This is a simple password manger (library) written in Go.
You can use it with Go, REST calls and any other language that links with C.

And the best of it all: **it's free** - BSD licensed - just don't sue me if you loose your passwords (no pun intended, ha!)


## Setup / Build

### Go

`go mod tidy` automatically fetches the necessary dependencies when you add the import statement to your code (see also: [Go's module support](https://go.dev/wiki/Modules#how-to-use-modules)):
```golang
import "github.com/image357/password"
```
Alternatively, use `go get` in your project to prefetch all dependencies:
```shell
go get -u github.com/image357/password/...@latest
```

### C/C++

You can use `cmake` to build and install the C interface library:
```shell
mkdir build; cd build
cmake -DCMAKE_INSTALL_PREFIX=/full/path/to/install/dir ..
cmake --build .
cmake --install .
```

Then, simply find the installed package in your `CMakeLists.txt`:
```cmake
find_package(password)
message(STATUS "${password_DLL_FILE}")

add_executable(main main.cpp)
target_link_libraries(main PRIVATE password::cinterface)
```

On Windows you will need MinGW to *build* the library.
This is because [cgo](https://go.dev/wiki/cgo) doesn't support any other compiler backend yet.
Once compiled, you can *use* the library with any compiler, though.
You can also have a look at the [release section](https://github.com/image357/password/releases) to see if there are any pre-built binaries for your platform.


## Usage

### Go
```golang
package main

import (
    "fmt"
    "github.com/image357/password"
    "github.com/image357/password/log"
    "github.com/image357/password/rest"
)

func main() {
    // create password with id
    password.Overwrite("myid", "mypassword", "storage_key")
    
    // get password with id
    pwd, _ := password.Get("myid", "storage_key")
    fmt.Println(pwd)
    
    // start a multi password rest service on localhost:8080
    rest.StartMultiService(
        ":8080", "/prefix", "storage_key",
        func(string, string, string, string) bool { return true },
    )
    
    // make logging more verbose
    log.Level(log.LevelDebug)
}
```

### C/C++:
```cpp
#include <stdio.h>
#include <password/cinterface.h>

bool callback(cchar_t *token, cchar_t *ip, cchar_t *resource, cchar_t *id) {
    return true;
}

int main() {
    // create password with id 
    CPWD__Overwrite("myid", "mypassword", "storage_key");
    
    // get password with id
    char buffer[256];
    CPWD__Get("myid", "storage_key", buffer, 256);
    printf("%s\n", buffer);
    
    // start a multi password rest service on localhost:8080
    CPWD__StartMultiService(":8080", "/prefix", "storage_key", callback);
    
    // make logging more verbose
    CPWD__LogLevel(CPWD__LevelDebug);
    
    return 0;
}
```

### Example Service
There is an example REST service, which you can use for testing.
Install it with:

```shell
go install github.com/image357/password/cmd/exampleservice@latest
```

Warning: do not use this service for production use-cases.
It doesn't have any access control and the storage key hides in [plain sight](./cmd/exampleservice/main.go)!

### REST
When you have your REST service running (see above) you can make calls with, e.g., python
```python
import requests

# create password with id
requests.put("http://localhost:8080/prefix/overwrite", json={"id": "myid", "password": "mypassword", "accessToken": "some_token"})

# get password with id
r = requests.get("http://localhost:8080/prefix/get", json={"id": "myid", "accessToken": "some_token"})
print(r.content)

# output: b'{"password":"mypassword"}'
```


# Details

### API
The REST API mirrors the [Go](./docs/password.md) and [C\C++](./docs/cinterface.md) interface.
This means that the signature is the same - with three exceptions:

1. For REST, you need an `accessToken`. For `Go/C/C++` you need a `storage_key`.
2. The C/C++ API accepts result pointers and returns status codes. The Go API directly returns values and errors.
3. The [simple service](./docs/rest.md#StartSimpleService) does not require "id" properties and will not bind to `Clean` and `List` calls.

Here is a full overview:

Overwrite:
```text
Go:   -> password.Overwrite(id string, password string, key string)
C/C++ -> CPWD__Overwrite(const char *id, const char *password, const char *key)
REST: -> (PUT) /prefix/overwrite

Example JSON for REST request:
{
    "accessToken": "my_token"
    "id": "my_id"
    "password": "my_password"
}

Return: {}
```

Get:
```text
Go:   -> password.Get(id string, key string)
C/C++ -> CPWD__Get(const char *id, const char *key, char *buffer, int length)
REST: -> (GET) /prefix/get

Example JSON for REST request:
{
    "accessToken": "my_token"
    "id": "my_id"
}

Return: {"password": "stored_password"}
```

Check:
```text
Go:   -> password.Get(id string, password string, key string)
C/C++ -> CPWD__Check(const char *id, const char *password, const char *key, bool *result)
REST: -> (GET) /prefix/check

Example JSON for REST request:
{
    "accessToken": "my_token"
    "id": "my_id"
    "password": "password_to_check"
}

Return: {"result": true/false}
```

Set:
```text
Go:   -> password.Set(id string, oldPassword string, newPassword string, key string)
C/C++ -> CPWD__Set(const char *id, const char *oldPassword, const char *newPassword, const char *key)
REST: -> (PUT) /prefix/set

Example JSON for REST request:
{
    "accessToken": "my_token"
    "id": "my_id"
    "oldPassword": "password1"
    "newPassword": "password2"
}

Return: {}
```

Unset:
```text
Go:   -> password.Unset(id string, password string, key string)
C/C++ -> CPWD__Unset(const char *id, const char *password, const char *key)
REST: -> (DELETE) /prefix/unset

Example JSON for REST request:
{
    "accessToken": "my_token"
    "id": "my_id"
    "password": "password_to_check"
}

Return: {}
```

Exists:
```text
Go:   -> password.Exists(id string)
C/C++ -> CPWD__Exists(const char *id, bool *result)
REST: -> (GET) /prefix/exists

Example JSON for REST request:
{
    "accessToken": "my_token"
    "id": "my_id"
}

Return: {"result": true/false}
```

List:
```text
Go:   -> password.List()
C/C++ -> CPWD__List(char *buffer, int length, const char *delim)
REST: -> (GET) /prefix/list

Example JSON for REST request:
{
    "accessToken": "my_token"
}

Return: {"ids": ["stored_id", "another_stored_id", ...]}
```

Delete:
```text
Go:   -> password.Delete(id string)
C/C++ -> CPWD__Delete(const char *id)
REST: -> (DELETE) /prefix/delete

Example JSON for REST request:
{
    "accessToken": "my_token"
    "id": "my_id"
}

Return: {}
```

Clean:
```text
Go:   -> password.Clean()
C/C++ -> CPWD__Clean()
REST: -> (DELETE) /prefix/clean

Example JSON for REST request:
{
    "accessToken": "my_token"
}

Return: {}
```

### Storage
Files and folders - it's that simple.
To make the storage backend cross-platform compatible, ids have the following constraints:

1. Forward- and backward-slashes are treated as the same character.
2. Upper- and lower-case characters are treated as the same character.

You can also switch to temporary (in-memory) storage or serialize to JSON (see below for full docs).

### Encryption
Yes, the usual - AES256, hashed secrets, etc.
For more info have a look at the [source code](./encryption.go).

### Documentation
For full documentation see: [docs](./docs/README.md)

There are some additional convenience functions in Go and C/C++ that control
storage path, logging, recovery mode and changing storage keys.
Additionally, you can just explore the source code with any godoc compatible IDE.

Check it out!
