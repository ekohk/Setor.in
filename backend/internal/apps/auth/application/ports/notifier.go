package ports

// Notifier abstracts the email service for usecases.
//
// Defining this interface here (instead of importing the email service
// directly) keeps the application layer free of infrastructure dependencies
// and lets us swap or mock notifications easily in tests.
type Notifier interface {
	// Collector application lifecycle
	CollectorApplicationSubmitted(to, fullName, businessName, address, licenseNo string)
	CollectorApplicationApproved(to, fullName, businessName string)
	CollectorApplicationRejected(to, fullName, businessName, reason string)

	// Account moderation
	AccountSuspended(to, fullName, reason string)
	AccountUnsuspended(to, fullName string)

	// Role / privileges
	RoleChanged(to, fullName, oldRole, newRole, reason string)

	// Security (reserved for future Keycloak event listener integration)
	PasswordChanged(to, fullName, changedAt, ip string)
	PasswordResetRequested(to, fullName, resetURL string)
	LoginNewDevice(to, fullName, loginAt, device, ip, location string)
}
