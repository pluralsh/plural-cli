package permissions

type Checker interface {
	MissingPermissions() ([]string, error)
}
