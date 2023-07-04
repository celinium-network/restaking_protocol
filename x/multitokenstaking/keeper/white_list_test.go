package keeper_test

import (
	"golang.org/x/exp/slices"
)

func (s *KeeperTestSuite) TestAddMTStakingDenom() {
	tests := []struct {
		describe     string
		existedDenom []string
		addedDenom   string
		successful   bool
	}{
		{
			describe:     "added in empty list",
			existedDenom: []string{},
			addedDenom:   "add",
			successful:   true,
		},

		{
			describe:     "added in no empty list",
			existedDenom: []string{"denom1", "denom2", "denom3"},
			addedDenom:   "denom4",
			successful:   true,
		},
		{
			describe:     "add existed denom",
			existedDenom: []string{"denom1", "denom2", "denom3"},
			addedDenom:   "denom1",
			successful:   false,
		},
	}

	for _, test := range tests {
		// init store
		s.SetupTest()

		for _, exist := range test.existedDenom {
			success := s.mtStakingKeeper.AddMTStakingDenom(s.ctx, exist)
			s.Require().True(success, "prepare existed denom failed, denom: ", exist)
		}

		added := s.mtStakingKeeper.AddMTStakingDenom(s.ctx, test.addedDenom)
		s.Require().Equal(test.successful, added, test.describe)
	}
}

func (s *KeeperTestSuite) TestGetMTStakingDenomWhiteList() {
	tests := []struct {
		describe     string
		existedDenom []string
		found        bool
	}{
		{
			describe:     "empty list",
			existedDenom: []string{},
			found:        false,
		},

		{
			describe:     "no empty list",
			existedDenom: []string{"denom1", "denom2", "denom3"},
			found:        true,
		},
	}

	for _, test := range tests {
		s.SetupTest()

		for _, exist := range test.existedDenom {
			success := s.mtStakingKeeper.AddMTStakingDenom(s.ctx, exist)
			s.Require().True(success, "prepare existed denom failed, denom: ", exist)
		}

		denomList, found := s.mtStakingKeeper.GetMTStakingDenomWhiteList(s.ctx)
		s.Require().Equal(test.found, found, test.describe)
		if found {
			slices.Equal(denomList.DenomList, test.existedDenom)
		}
	}
}

func (s *KeeperTestSuite) TestRemoveMTStakingDenom() {
	tests := []struct {
		describe     string
		existedDenom []string
		removedDenom string
		successful   bool
	}{
		{
			describe:     "remove in empty list",
			existedDenom: []string{},
			removedDenom: "add",
			successful:   false,
		},

		{
			describe:     "added in no empty list",
			existedDenom: []string{"denom1", "denom2", "denom3"},
			removedDenom: "denom3",
			successful:   true,
		},
		{
			describe:     "remove no existed denom",
			existedDenom: []string{"denom1", "denom2", "denom3"},
			removedDenom: "denom4",
			successful:   false,
		},
	}

	for _, test := range tests {
		s.SetupTest()

		for _, exist := range test.existedDenom {
			success := s.mtStakingKeeper.AddMTStakingDenom(s.ctx, exist)
			s.Require().True(success, "prepare existed denom failed, denom: ", exist)
		}

		successful := s.mtStakingKeeper.RemoveMTStakingDenom(s.ctx, test.removedDenom)
		s.Require().Equal(test.successful, successful, test.describe)
		if test.successful {
			list, _ := s.mtStakingKeeper.GetMTStakingDenomWhiteList(s.ctx)
			s.Require().False(slices.Contains(list.DenomList, test.removedDenom), test.describe)
		}
	}
}
