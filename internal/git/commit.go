package git

type Commit struct {
	// tree         ObjectPointer
	// parents      []ObjectPointer
	// extraHeaders []struct {
	// 	key   string
	// 	value string
	// }
	// message string
	buffer []byte
}

func (c *Commit) RawGitBuffer() ([]byte, error) {
	return c.buffer, nil
}
