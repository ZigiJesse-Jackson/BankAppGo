package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	pwd := RandomString(6)
	hashed_pwd_1, err := HashPassword(pwd)
	require.NoError(t, err)
	require.NotEmpty(t, hashed_pwd_1)

	err = CheckPassword(pwd, hashed_pwd_1)
	require.NoError(t, err)

	wrong_pwd := RandomString(6)
	err = CheckPassword(wrong_pwd, hashed_pwd_1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	hashed_pwd_2, err := HashPassword(pwd)
	require.NoError(t, err)
	require.NotEmpty(t, hashed_pwd_2)
	require.NotEqual(t, hashed_pwd_1, hashed_pwd_2)

}
