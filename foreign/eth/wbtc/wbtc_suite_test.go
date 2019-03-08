package wbtc_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWbtc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Wbtc Suite")
}
