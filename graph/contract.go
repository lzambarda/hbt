package graph

type Graph interface {
	Track(id, wd, cmd string)
	Hint(id, wd string) string
	End(id string)
	Save(filePath string) error
	Load(filePath string) error
}
