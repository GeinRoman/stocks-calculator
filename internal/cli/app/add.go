package app

type AddOptions struct {
	Stock bool
	Group bool
}

func Add(options *AddOptions, args []string) (string, error) {
	return "Be sure, something was addded", nil
}
