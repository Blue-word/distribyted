package loader

type Loader interface {
	ListMagnets(user string) (map[string][]string, error)
	ListTorrentPaths(user string) (map[string][]string, error)
}

type LoaderAdder interface {
	Loader

	RemoveFromHash(r, h string) (bool, error)
	AddMagnet(r, m, user string) error
}
