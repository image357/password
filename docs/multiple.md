# Handling of multiple password manager instances

The C-style and Go API exposes two functions for handling multiple instances of password managers via string identifiers:

* `SetDefaultManager` / `CPWD__SetDefaultManager`
* `RegisterDefaultManager` / `CPWD__RegisterDefaultManger`

They enable you to configure each instance step by step via the appropriate function calls (e.g. storage backend).
Once configuration is done, you can simply push the current default manager onto the "named stack" via `RegisterDefaultManager`.
In Go, you can access this stack via the global variable `Managers`. In C/C++ you have to retrieve managers indirectly via `CPWD__SetDefaultManager`.
All standard interface calls will always route to the current default manger.

This mechanism is also useful for REST service creation.
The current default manager can be pushed onto the stack and be used in subsequent REST requests / callbacks.

```golang
package main 

import "github.com/image357/password"
import "github.com/image357/password/rest"

func main() {
    password.SetStorePath("some/path")
    rest.StartSimpleService(...)
    password.RegisterDefaultManager("rest service 1")

    password.SetStorePath("another/path")
    rest.StartSimpleService(...)
    password.RegisterDefaultManager("rest service 2")
}
```

When working with the C API in a multi manager setup, you can use the manager specific function set:
```cpp
CPWD__mOverwrite("some_manager", ...);
CPWD__mGet("some_manager", ...);
CPWD__mCheck("some_manager", ...);
CPWD__mSet("some_manager", ...);
CPWD__mUnset("some_manager", ...);
CPWD__mExists("some_manager", ...);
CPWD__mList("some_manager", ...);
CPWD__mDelete("some_manager", ...);
CPWD__mClean("some_manager", ...);
CPWD__mRewriteKey("some_manager", ...);
```
For instance, the default manager can be referenced via the `"default"` identifier string.
