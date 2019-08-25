package gonetmon

import (
	"github.com/sirupsen/logrus"
	"os"
)

var log = logrus.New()
var config *configuration

func init() {

	// Must be root or sudo
	if os.Geteuid() != 0 {
		log.Error("Geteuid is not 0 : not running with elevated privileges.\n" +
			"You must run this program with elevated privileges in order to capture network traffic. Try running with sudo.")
	}

	// Load default parameters
	config = LoadParams()
}
