package app

type LoginOptions struct {
	NewAccount bool
}

func Login(user string, options *LoginOptions) (string, error) {
	return "logged in successfully", nil
}
