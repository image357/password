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
