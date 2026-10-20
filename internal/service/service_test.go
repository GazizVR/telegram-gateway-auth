package service

import "testing" 

const code = "123456"

func TestSignHash(t *testing.T) {
    key := []byte("qfMszzVLzZxYPA6nqXix9OgmTf1xFx+7+8h8f9PXGHI=")
    hash := signHash([]byte(code),key)
    if len(hash) < 1 {
        t.Fatal("hash must not be empty")
    }
}

var hash = []byte{
	251, 111, 227, 5, 149, 211, 228, 99,
	75, 127, 90, 147, 225, 138, 92, 179,
	71, 176, 156, 187, 47, 205, 182, 163,
	140, 124, 202, 116, 215, 147, 200, 201,
}
   
var key = []byte("qfMszzVLzZxYPA6nqXix9OgmTf1xFx+7+8h8f9PXGHI=")  

func TestVerifyHash(t *testing.T) {
    if !verifyHash([]byte(code),hash,key) {
        t.Fatal("hash was supposed to be valid")
    }
}

func TestVerifyHashInvalid(t *testing.T) {
    code := "654321"
    if verifyHash([]byte(code),hash,key) {
        t.Fatal("hash was supposed to be invalid")
    }
}