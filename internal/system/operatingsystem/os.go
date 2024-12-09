package operatingsystem

type OS struct {
	ID      string
	Name    string
	Version string
}

func Info() (*OS, error) {
	return platformInfo()
}
