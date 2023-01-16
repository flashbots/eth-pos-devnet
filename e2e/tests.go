package e2e

type TestSpec struct {
	Name        string
	Run			func()
}

// TODO: split into only builder and devnet tests
var Tests = []TestSpec{
	{
		Name: "dummy test",
	},
}

func getTests() []TestSpec {
	return Tests
}