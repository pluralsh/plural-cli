package permissions

type nullChecker struct {
}

func NullChecker() *nullChecker { return &nullChecker{} }

func (*nullChecker) MissingPermissions() ([]string, error) {
	return []string{}, nil
}
