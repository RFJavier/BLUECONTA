package backend

// LicenseManager is an extension point for a future licensing module.
type LicenseManager struct {
	status string
}

func NewLicenseManager() *LicenseManager {
	return &LicenseManager{status: "unlicensed-dev"}
}

func (l *LicenseManager) AllowsWriteOperations() bool {
	return true
}

func (l *LicenseManager) Status() string {
	return l.status
}
