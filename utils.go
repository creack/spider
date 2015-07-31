package spider

// TrimBytePrefix trims the given prefix.
func TrimBytePrefix(str string, prefix byte) string {
	if len(str) > 0 && str[0] == prefix {
		str = str[1:]
	}
	return str
}

// TrimByteSuffix trims the given suffix.
func TrimByteSuffix(str string, suffix byte) string {
	if len(str) > 0 && str[len(str)-1] == suffix {
		str = str[:len(str)]
	}
	return str
}

// TrimSlash trims '/' from the given string.
func TrimSlash(str string) string {
	return TrimByteSuffix(TrimBytePrefix(str, '/'), '/')
}
