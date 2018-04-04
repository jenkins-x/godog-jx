package godog_jx

import (
	"testing"

	"flag"
	"fmt"
	"github.com/DATA-DOG/godog"
	"os"
	"time"
)

var (
	featureFlag = ""
)

func init() {
	flag.StringVar(&featureFlag, "godog.feature", "import.feature", "The godog feature to run")
	flag.Parse()
}

/*
var (
	featureFlag = flag.String("godog.feature", "importurl.feature", "The godog feature to run")
)
*/

func TestMain(t *testing.M) {
	testing.CoverMode()
	flag.Parse()

	fmt.Printf("Running feature %s\n", featureFlag)
	status := godog.RunWithOptions("godog", findSuite, godog.Options{
		Format:    "progress",
		Paths:     []string{featureFlag},
		Randomize: time.Now().UTC().UnixNano(), // randomize scenario execution order
	})

	if st := t.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func findSuite(s *godog.Suite) {
	switch featureFlag {
	case "import.feature":
		ImportFeatureContext(s)
	case "importurl.feature":
		ImporturlFeatureContext(s)
	case "spring.feature":
		SpringFeatureContext(s)
	default:
		ImportFeatureContext(s)
	}
}