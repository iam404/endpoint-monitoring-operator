package notifier

type Notifier interface {
	SendAlert(status string, msg string) error
}
