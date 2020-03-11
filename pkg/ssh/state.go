package ssh

import (
	"github.com/kubernetes-simulator/simulator/pkg/childminder"
	"github.com/kubernetes-simulator/simulator/pkg/util"
	"github.com/pkg/errors"
	"os"
	"strings"
)

// SSHState provides methods for storing or retrieving state about a user and
// their cluster
type SSHState interface {
	GetSSHKeyPair() (*KeyPair, error)
	SaveSSHConfig(config string) error
	GetSSHConfig() (*string, error)
}

// LocalState is the default State provider and persists all state into the
// local ~/.kubesim directory
type LocalState struct{}

// GetSSHKeyPair returns an existing SSH keypair or creates one locally
func (ls LocalState) GetSSHKeyPair() (*KeyPair, error) {
	abspath, err := util.ExpandTilde(PrivateKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "Error resolving key path")
	}

	exists, err := util.FileExists(*abspath)
	if err != nil {
		return nil, errors.Wrap(err, "Error checking if key already exists")
	}

	// key already exists return it
	if exists {
		return ls.getSSHKeyPair()
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting process working directory")
	}

	cm := childminder.NewChildMinder(nil, wd, os.Environ(), "ssh-keygen",
		"-f", *abspath, "-t", "rsa", "-C",
		"simulator-key", "-N", "")
	_, _, err = cm.RunSilently()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating keypair")
	}

	return ls.getSSHKeyPair()
}

// GetSSHKeyPair retieves
func (ls LocalState) getSSHKeyPair() (*KeyPair, error) {
	publicKeyPath, err := util.ExpandTilde(PublicKeyPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error resolving %s", PublicKeyPath)
	}

	publickey, err := util.Slurp(*publicKeyPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading %s", PublicKeyPath)
	}

	privateKeyPath, err := util.ExpandTilde(PrivateKeyPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error resolving %s", PrivateKeyPath)
	}

	privatekey, err := util.Slurp(*privateKeyPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading %s", PublicKeyPath)
	}
	ret := KeyPair{
		PublicKey:  SSHPublicKey(strings.Trim(*publickey, "\n")),
		PrivateKey: SSHPrivateKey(*privatekey),
	}
	return &ret, nil
}

// SaveSSHConfig saves the config supplied to the local ~/.ssh directory
func (ls LocalState) SaveSSHConfig(config string) error {
	abspath, err := util.ExpandTilde(SSHConfigPath)
	if err != nil {
		return errors.Wrap(err, "Error resolving SSH config path")
	}

	err = util.OverwriteFile(*abspath, config)
	if err != nil {
		return errors.Wrap(err, "Error overwriting SSH config")
	}

	return nil
}

// GetSSHConfig reads the config from the local ~/.ssh directory
func (ls LocalState) GetSSHConfig() (*string, error) {
	abspath, err := util.ExpandTilde(SSHConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error resolving %s", SSHConfigPath)
	}

	config, err := util.Slurp(*abspath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading %s", SSHConfigPath)
	}
	return config, nil
}
