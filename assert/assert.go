package assert

// TestingT is an interface wrapper around *testing.T
type TestingT interface {
	Errorf(format string, args ...interface{})
}

func Equal(t TestingT, expected, actual interface{}) bool {
	if expected != actual {
		t.Errorf("expected %s doesn't match %s", expected, actual)
		return false
	}
	return true
}
