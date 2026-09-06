package password

import "golang.org/x/crypto/bcrypt"

const DefaultCost = 12

func Hash(value string) (string, error) { return HashWithCost(value, DefaultCost) }
func HashWithCost(value string, cost int) (string, error) {
	result, err := bcrypt.GenerateFromPassword([]byte(value), cost)
	return string(result), err
}
func Verify(hash, value string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(value)) == nil
}
