package git

type Tree struct {
	// entries []ObjectPointer
	buffer []byte
}

func (t *Tree) RawGitBuffer() ([]byte, error) {
	return t.buffer, nil
}
