package testcommon

type MockStrGen struct{}

const TestStr = "short-hash"

func (m *MockStrGen) GetRandomString(n int) string {
	return TestStr
}
