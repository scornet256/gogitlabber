package logging

var applicationName = ""

// Sets the application name prefix used in the logoutput. Example:
// <applicationName> | ...
func SetAppName(name string) {
	applicationName = name
}

// Returns the logging prefix name as string.
func GetAppName() string {
	return applicationName
}

