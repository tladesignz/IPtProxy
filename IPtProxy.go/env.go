package IPtProxy

import (
	"errors"
	"io/fs"
	"os"
)

// Hack: Set some environment variables that are either
// required, or values that we want. Have to do this here, since we can only
// launch this in a thread and the manipulation of environment variables
// from within an iOS app won't end up in goptlib properly.
//
// Note: This might be called multiple times when using different functions here,
// but that doesn't necessarily mean, that the values set are independent each
// time this is called. It's still the ENVIRONMENT, we're changing here, so there might
// be race conditions.
func fixEnv(stateDir string) {
	info, err := os.Stat(stateDir)

	// If dir does not exist, try to create it.
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(stateDir, 0700)

		if err == nil {
			info, err = os.Stat(stateDir)
		}
	}

	// If it is not a dir, panic.
	if err == nil && !info.IsDir() {
		err = fs.ErrInvalid
	}

	// Create a file within dir to test writability.
	if err == nil {
		tempFile := stateDir + "/.iptproxy-writetest"
		var file *os.File
		file, err = os.Create(tempFile)

		// Remove the test file again.
		if err == nil {
			file.Close()

			err = os.Remove(tempFile)
		}
	}

	if err != nil {
		panic("Error with stateDir directory \"" + stateDir + "\":\n" +
			"  " + err.Error() + "\n" +
			"  stateDir needs to be set to a writable directory.\n" +
			"  Use an app-private directory to avoid information leaks.\n" +
			"  Use a non-temporary directory to allow reuse of potentially stored state.")
	}

	_ = os.Setenv("TOR_PT_CLIENT_TRANSPORTS", "meek_lite,obfs2,obfs3,obfs4,scramblesuit,webtunnel,snowflake")
	_ = os.Setenv("TOR_PT_MANAGED_TRANSPORT_VER", "1")
	_ = os.Setenv("TOR_PT_STATE_LOCATION", stateDir)
}
