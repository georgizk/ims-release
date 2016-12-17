package assert

import "runtime"

// TestingT is an interface wrapper around *testing.T
type TestingT interface {
	Errorf(format string, args ...interface{})
}

func NotEqual(t TestingT, unexpected, actual interface{}) bool {
	if unexpected == actual {
		_, file, no, _ := runtime.Caller(1)
		t.Errorf("%s#%d: did not expect %s to match", file, no, unexpected)
		return false
	}
	return true
}

func Equal(t TestingT, expected, actual interface{}) bool {
	if expected != actual {
	    _, file, no, _ := runtime.Caller(1)
        t.Errorf("%s#%d: expected %s doesn't match %s", file, no, expected, actual)
		return false
	}
	return true
}
