package file

type File struct {
	path     string
	checksum string
}

func New(p string) (*File, error) {
	if err := Exists(p); err != nil {
		return &File{}, err
	}
	c, err := Checksum(p)
	if err != nil {
		return &File{}, err
	}

	return &File{
		path:     p,
		checksum: c,
	}, nil
}

func (f *File) Checksum() string {
	return f.checksum
}

// CompareChecksum returns true only if the sha1 is not equal to the stored value
func (f *File) CompareChecksum() (bool, error) {
	new, err := Checksum(f.path)
	if err != nil {
		return false, err
	}
	if new == f.checksum {
		return false, nil
	}

	return true, nil
}

func (f *File) Path() string {
	return f.path
}
