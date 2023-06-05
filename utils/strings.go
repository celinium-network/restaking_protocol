package utils

func BytesLengthPrefix(bz []byte) []byte {
	bzLen := len(bz)
	if bzLen == 0 {
		return bz
	}
	return append([]byte{byte(bzLen)}, bz...)
}
