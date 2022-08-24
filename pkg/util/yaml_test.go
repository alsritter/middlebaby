package util

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"testing"
)

func TestErrorPackage(t *testing.T) {
	var result error

	result = multierror.Append(result, fmt.Errorf("这是错误01"))
	result = multierror.Append(result, fmt.Errorf("这是错误02"))
	result = multierror.Append(result, fmt.Errorf("这是错误03"))
	t.Log(result)
}
