[WIP] few tools to manuplate PAN numbers

# Work In Progress

This software is still work in progress, the API could change undefinitely.

# Synopsis

```sh
go get github.com/raohwork/creditcard
```

```golang
if err := creditcard.FromDashed("1234-5678-9012-3456").Validate(); err != nil {
    log.Fatal("Your credit card number is invalid")
}
```

# License

MPL 2.0, see LICENSE.txt.
