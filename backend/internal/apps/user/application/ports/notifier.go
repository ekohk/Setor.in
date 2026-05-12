package ports

// Notifier abstracts welcome / first-sync notifications for the user module.
type Notifier interface {
	Welcome(to, fullName string)
}
