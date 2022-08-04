package mqtt

// 1.5.5 Variable Byte Integer
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
func NewVarInt(x uint) []byte {
	result := make([]byte, 0, 4) // max four
	if x == 0 {
		result = append(result, 0)
		return result
	}
	for x > 0 {
		encodedByte := byte(x % 128)
		x = x / 128
		if x > 0 {
			encodedByte = encodedByte | 128
		}
		result = append(result, encodedByte)
	}
	return result
}
