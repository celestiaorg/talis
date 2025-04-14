package provisioner

// MockSSHChecker implements SSHChecker interface for testing
type MockSSHChecker struct {
	// WaitForSSHFunc allows customizing the behavior of WaitForSSH
	WaitForSSHFunc func(host string) error
}

// WaitForSSH implements SSHChecker.WaitForSSH
func (m *MockSSHChecker) WaitForSSH(host string) error {
	if m.WaitForSSHFunc != nil {
		return m.WaitForSSHFunc(host)
	}
	return nil // By default, pretend SSH is immediately available
}
