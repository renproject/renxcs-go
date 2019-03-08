package renxcs_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRenxcsGo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RenxcsGo Suite")
}
