package model

func ValidWindowState(state string) bool {
	return state == WindowPlanned || state == WindowActive || state == WindowCompleted || state == WindowFailed
}
func ValidCommandState(state string) bool {
	return state == CommandQueued || state == CommandSending || state == CommandSent || state == CommandFailed
}
func IsTerminalWindow(state string) bool {
	return state == WindowCompleted || state == WindowFailed || state == WindowCancelled
}
func IsTerminalCommand(state string) bool { return state == CommandSent || state == CommandFailed }
