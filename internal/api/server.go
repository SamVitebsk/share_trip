package api

type Server struct {
	tripHandler  *TripHandler
	readyHandler *ReadyHandler
}

func NewServer(tripHandler *TripHandler, readyHandler *ReadyHandler) *Server {
	return &Server{
		tripHandler:  tripHandler,
		readyHandler: readyHandler,
	}
}
