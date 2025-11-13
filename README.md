Errific
=========

<img src="./errific.png" width="340" alt="Errific Art"><br>

Super simple error strings in Go with caller prefix|suffix metadata, clean error wrapping, and helpful formatting methods.

### Using [New](https://github.com/leefernandes/errific/blob/main/error.go#L25) and [Errorf](https://github.com/leefernandes/errific/blob/main/error.go#L39) to format an error message and handle error types.
```go
var (
	ErrRegisterPet  errific.Err = "error registering pet"
	ErrValidateKind errific.Err = "only cats are allowed, cannot register '%s'"
)

func main() {
	if err := registerPet("hamster"); err != nil {
		switch {
		case errors.Is(err, ErrValidateKind): // 400 errors
			fmt.Println(http.StatusBadRequest, err)

		default: // 500 errors
			fmt.Println(http.StatusInternalServerError, err)
		}
	}
}

func registerPet(kind string) error {
	if err := validateKind(kind); err != nil {
		return ErrRegisterPet.New(err)
	}
	return nil
}

func validateKind(kind string) error {
	if kind != "cat" {
		return ErrValidateKind.Errorf(kind)
	}
	return nil
}
```
```shell
400 error registering pet [/tmp/sandbox4095574913/prog.go:30.registerPet]
only cats are allowed, cannot register 'hamster' [/tmp/sandbox4095574913/prog.go:37.validateKind]
```


Try it on the <a href="https://go.dev/play/p/N7asgc_1i-J"><img src="./gopher.png" height="14px" /></a> [playground](https://go.dev/play/p/N7asgc_1i-J)!

More to come!  In the meantime look at the [example tests](https://github.com/leefernandes/errific/tree/main/examples).
