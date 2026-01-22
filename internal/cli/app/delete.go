package app

type DeleteOptions struct {
	Stock bool
	Group bool
	All bool
}

func Delete(options *DeleteOptions, args []string) (string, error) {
	return "deleting everything you want", nil
}
