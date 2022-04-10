package shaders

type Config struct {
	RootDir  string             `json:"-"`
	Include  []string           `json:"include"`
	Programs map[string]Program `json:"programs"`
}

type Program struct {
	Vertex   string   `json:"vertex"`
	Fragment string   `json:"fragment"`
	Geometry string   `json:"geometry"`
	Compute  string   `json:"compute"`
	Flags    []string `json:"flags"`
}
