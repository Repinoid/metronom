package privacy

import (
	"crypto/md5"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptB2B(t *testing.T) {

	testingString := "\t во поле берёзка стояла, в натуре\n"

	passwordKey := MakeHash(nil, []byte(testingString), []byte(testingString)) // let it be

	keyB16 := md5.Sum(passwordKey)
	require.Equal(t, len(keyB16), 16)

	keyB := keyB16[:]
	coded, err := EncryptB2B([]byte(testingString), keyB)
	require.NoError(t, err)

	telo, err := DecryptB2B(coded, keyB)
	require.NoError(t, err)

	require.Equal(t, testingString, string(telo))

}
