package domain

const (
	InsufficientPrivilegesError string = "user has insufficient privileges"
	BadUserCredentialError      string = "bad user credentials"
)

const (
	InvalidEmail string = "email is invalid"
)

const (
	AnotherOpenedReceptionError   string = "pvz has another receptions opened"
	AllReceptionsAreClosed        string = "all receptions are closed at this pvz"
	ReceptionIsAlreadyClosedError string = "reception is already closed"
	ReceptionIsEmptyError         string = "no products in reception"
)

const (
	InvalidIdStateError string = "id value was invalid"
)

const (
	UnknownCityError            string = "unknown city"
	UnknownProductCategoryError string = "unknown product category"
	PVZDoesNotExistError        string = "pvz was not found"
	UnknownRoleNameError        string = "unknown user role"
)

func IsAccessError(err error) bool {
	switch err.Error() {
	case InsufficientPrivilegesError, BadUserCredentialError:
		return true
	default:
		return false
	}
}
