package config

func needWriteFile(filename string) (bool, error) {
	if filename == "" {
		return false, nil
	}

	return true, nil
}
