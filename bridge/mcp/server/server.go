package server

type Server struct {
	Name string
}

func NewServer(name string) *Server {
	return &Server{Name: name}
}
