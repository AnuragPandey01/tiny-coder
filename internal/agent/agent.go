package agent

type Agent interface {
	Run(string) <-chan string
}
