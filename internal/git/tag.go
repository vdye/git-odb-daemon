package git

type Tag struct {
	// target       ObjectPointer
	// extraHeaders []struct {
	// 	key   string
	// 	value string
	// }
	// message string
	buffer []byte
}

func (t *Tag) RawGitBuffer() ([]byte, error) {
	return t.buffer, nil
}
