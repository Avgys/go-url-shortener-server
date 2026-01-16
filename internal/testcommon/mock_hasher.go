package testcommon

type MockHasher struct{}

const ShortHash = "short-hash"

func (m *MockHasher) GetHash(input string) string {
	return ShortHash
}
