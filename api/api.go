package api

type Init struct {
	Node    string
	Streams []string
}

type LogEvent struct {
	Stream string
	Msg    string
}
