package nex

import "crypto/md5"


// MD5Hash returns the MD5 hash of the input
func MD5Hash(text []byte) []byte {
	hasher := md5.New()
	hasher.Write(text)
	return hasher.Sum(nil)
}