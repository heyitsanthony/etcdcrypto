package etcdcrypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	for i := 0; i < 32; i++ {
		key[i] = byte(i)
	}
	c, err := NewAESCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	ss := []string{
		"0123456789012345",
		string(key),
		"",
		"a",
		"bb",
		"ccc",
		"AAAAAaaAAAAAAAAAAA",
	}
	for i, s := range ss {
		v, derr := c.Decrypt(c.Encrypt([]byte(s)))
		if derr != nil {
			t.Errorf("#%d: decrypt failed (%v)", i, derr)
			continue
		}
		if s != string(v) {
			t.Errorf("#%d: got %q, expected %q\n", v, s)
			continue
		}
	}
}
