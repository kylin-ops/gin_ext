package validate

type ValidatorFun func() error

func Validator(fns ...ValidatorFun) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
