package zbtc_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestZbtc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zbtc Suite")
}
