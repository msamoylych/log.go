package api

type Init struct {
	Code string
	Streams []string
}

type LogEvent struct {
	Code byte
	Msg string
}
