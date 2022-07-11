package commit

type Commit struct {
	Message string
	Hash    string
	Author  string
}

// Source returns an ordered list of all following commits to a provided lastHash from a repository
type Source interface {
	Commits(lastHash string) ([]Commit, error)
}
