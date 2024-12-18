# Recovery

You can enable storage of recovery key files via `password.EnableRecovery` (or the corresponding C-API function).
Once enabled, all password write operations will store another file alongside the original password file.
If your password file is called `myid.pwd`, then the recovery key file will be `myid.recovery.pwd`.
It contains the encrypted storage key.
The encryption mechanism is the same as for regular passwords, but here the storage key is encrypted with the recovery key.
Hence, use regular `Get` operations of the standard API to access your lost storage key.

There is also a small helper tool under [/cmd/recovery](../cmd/recovery), which you can install via
```shell
go install https://github.com/image357/password/cmd/recovery@latest
```
Once installed, just point the executable towards anyone file (recovery or original password file):
```shell
# print myid password
recovery /full/path/to/myid.pwd RECOVERY_KEY

# print again
recovery /full/path/to/myid.recovery.pwd RECOVERY_KEY
```

You can always implement your own recovery- or multi-key protocol on top of the current API by simply encrypting the storage key.
However, this mechanism is in particular useful in combination with the available rest services, since you cannot alter the usage scheme of the storage key within the service.
For instance, the [exampleservice](../cmd/exampleservice) will write recovery key files by default.

If you know the storage key you can also use the encryption helper tools:
```shell
go install https://github.com/image357/password/cmd/encrypt@latest
go install https://github.com/image357/password/cmd/decrypt@latest

encrypt <file> <key>
decrypt <file> <key>
```
This will print the encrypted/decrypted contents to stdout.
