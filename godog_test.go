package godog_jx

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
)

var (
	featureFlag = ""
)

var opt = godog.Options{
	Output: colors.Colored(os.Stdout),
}

func init() {
	flag.StringVar(&featureFlag, "godog.feature", "import.feature", "The godog feature to run")
	godog.BindFlags("godog.", flag.CommandLine, &opt)
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
	status := godog.RunWithOptions("godog", FindSuite, godog.Options{
		Format:    "progress",
		Paths:     []string{featureFlag},
		Randomize: time.Now().UTC().UnixNano(), // randomize scenario execution order
		Concurrency: 6,
	})

	if st := t.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func FindSuite(s *godog.Suite) {
	switch featureFlag {
	case "import.feature":
		ImportFeatureContext(s)
	case "importurl.feature":
		ImporturlFeatureContext(s)
	case "spring.feature":
		SpringFeatureContext(s)
	case "quickstart-android-quickstart.feature":
		AndroidQuickstartFeatureContext(s)
	case "quickstart-angular-io-quickstart.feature":
		AngularIoQuickstartFeatureContext(s)
	case "quickstart-aspnet-app.feature":
		AspnetAppFeatureContext(s)
	case "quickstart-golang-http.feature":
		GolangHTTPFeatureContext(s)
	case "quickstart-node-http.feature":
		NodeHTTPFeatureContext(s)
	case "quickstart-rust-http.feature":
		RustHTTPFeatureContext(s)
	case "quickstart-python-http.feature":
		PythonHTTPFeatureContext(s)
	default:
		ImportFeatureContext(s)
	}
}
