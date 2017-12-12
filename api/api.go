package api

type Init struct {
	NodeName string
	Streams []string
}

type LogEvent struct {
	Code byte
	Msg string
}
