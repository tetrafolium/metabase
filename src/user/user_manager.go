package user

// Manager is an interface for getting/creating user information.
type Manager interface {
	GetID(service string, userID string) (uint64, error)
	CreateID(service string, userID string) (uint64, error)
}
