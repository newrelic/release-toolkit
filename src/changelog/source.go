package changelog

type Source interface {
	Changelog() (*Changelog, error)
}
