package godog_jx

import (
	"testing"

	"github.com/DATA-DOG/godog"
	"time"
	"os"
	"flag"
	"fmt"
)

var (
	featureFlag = flag.String("godog.feature", "importurl.feature", "The godog feature to run")
)

func TestMain(t *testing.M) {
	testing.CoverMode()
	flag.Parse()

	fmt.Printf("Running feature %s\n", *featureFlag)
	status := godog.RunWithOptions("godog", func(s *godog.Suite) {
		ImporturlFeatureContext(s)
	}, godog.Options{
		Format:    "progress",
		Paths:     []string{*featureFlag},
		Randomize: time.Now().UTC().UnixNano(), // randomize scenario execution order
	})

	if st := t.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
