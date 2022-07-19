package password

import (
	"testing"

	"github.com/milhamh95/simplebank/pkg/random"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	password := random.RandomString(6)

	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	t.Run("success", func(t *testing.T) {
		err = CheckPassword(password, hashedPassword)
		require.NoError(t, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		wrongPassword := random.RandomString(6)
		err = CheckPassword(wrongPassword, hashedPassword)
		require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
	})

	t.Run("hashed password is not equal", func(t *testing.T) {
		hashedPassword2, err := HashPassword(password)
		require.NoError(t, err)
		require.NotEmpty(t, hashedPassword2)
		require.NotEqual(t, hashedPassword, hashedPassword2)
	})
}
