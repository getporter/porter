package porter

// CNABProvider
type CNABProvider interface {
	Install() error
	//Upgrade() error
	//Uninstall() error
}
