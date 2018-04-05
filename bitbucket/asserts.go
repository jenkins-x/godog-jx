package bitbucket

import (
	"fmt"
	"os"
	"strings"

	"github.com/stretchr/testify/assert"
)

// ErrorSlice stores all of the failures
type ErrorSlice struct {
	Errors []string
}

func CreateErrorSlice() *ErrorSlice {
	return &ErrorSlice{
		Errors: []string{},
	}
}

func (t *ErrorSlice) Errorf(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	t.Errors = append(t.Errors, text)
}

// Error returns the error for this ErrorSlice or null
// if there are none yet
func (e *ErrorSlice) Error() error {
	if e.Errors == nil || len(e.Errors) == 0 {
		return nil
	}
	message := strings.Join(e.Errors, "\n")
	return fmt.Errorf("Assertions failed: %s", message)
}

func CreateAssert(t *ErrorSlice) *assert.Assertions {
	return assert.New(t)
}

func AssertFileDoesNotExist(path string) error {
	if _, err := os.Stat("/path/to/whatever"); os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("The path %s exists!", path)
}
