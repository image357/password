# Handling of multiple password manager instances

The C-style and Go API exposes two functions for handling multiple instances of password managers via string identifiers:

* `SetDefaultManager` / `CPWD__SetDefaultManager`
* `RegisterDefaultManager` / `CPWD__RegisterDefaultManger`

They enable you to configure each instance step by step via the appropriate function calls to, e.g., their storage backend.
Once configuration is done, you can simply push the current default password manager onto the "stack" via `RegisterDefaultManager`.
In Go, you can access this stack via the global variable `Managers`. In C/C++ you have to retrieve managers indirectly via `CPWD__SetDefaultManager`.
All standard interface calls will always route to the current default manger.

This mechanism is also used for REST service creation.
The current default manager will be pushed onto the stack and used in subsequent REST requests / callbacks.
