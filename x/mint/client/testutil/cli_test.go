package testutil

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/testutil/network"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
