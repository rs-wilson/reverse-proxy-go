package mocks

// MockConfig allows for control over a testable mock config
type MockConfig struct {
	UserCheck   bool
	PassCheck   bool
	UserAllowed bool
	Address     string
}

func (me *MockConfig) CheckUsername(string) bool {
	return me.UserCheck
}

func (me *MockConfig) CheckPassword(string, string) bool {
	return me.PassCheck
}

func (me *MockConfig) IsUserAllowed(string, string) bool {
	return me.UserAllowed
}

func (me *MockConfig) GetAddress(string) (string, bool) {
	return me.Address, true
}

// MockAuthProvider allows for control over a testable auth provider
type MockAuthProvider struct {
	Valid bool
}

func (me *MockAuthProvider) GetToken(string) (string, error) {
	return "", nil
}

func (me *MockAuthProvider) ValidationToken(string) (string, bool) {
	return "bob", me.Valid
}
