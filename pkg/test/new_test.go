/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func TestNew(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()

	t.Run("new-blank-username", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()

		// create user
		userObject := user.NewUsers(mockClient, ens, logger)
		_, _, _, _, _, err := userObject.CreateNewUserV2("", "password1", "", "", tm)
		if !errors.Is(err, user.ErrBlankUsername) {
			t.Fatal(err)
		}
	})

	t.Run("new-user", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()

		// create user
		userObject := user.NewUsers(mockClient, ens, logger)
		_, _, _, _, _, err := userObject.CreateNewUserV2("user1", "password1", "", "", tm)
		if err != nil && !errors.Is(err, user.ErrPasswordTooSmall) {
			t.Fatal(err)
		}

		_, mnemonic, _, _, ui, err := userObject.CreateNewUserV2("user1", "password1twelve", "", "", tm)
		if err != nil {
			t.Fatal(err)
		}

		_, _, _, _, _, err = userObject.CreateNewUserV2("user1", "password1twelve", "", "", tm)
		if !errors.Is(err, user.ErrUserAlreadyPresent) {
			t.Fatal(err)
		}

		// validate user
		if !userObject.IsUsernameAvailableV2(ui.GetUserName()) {
			t.Fatalf("user not created")
		}
		if !userObject.IsUserNameLoggedIn(ui.GetUserName()) {
			t.Fatalf("user not loggin in")
		}
		if ui == nil {
			t.Fatalf("invalid user info")
		}
		if ui.GetUserName() != "user1" {
			t.Fatalf("invalid user name")
		}
		if ui.GetFeed() == nil || ui.GetAccount() == nil || ui.GetPod() == nil {
			t.Fatalf("invalid feed, account or pod")
		}
		err = ui.GetAccount().GetWallet().IsValidMnemonic(mnemonic)
		if err != nil {
			t.Fatalf("invalid mnemonic")
		}
	})

	t.Run("new-user-multi-cred", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		user1 := "multicredtester"
		//create user
		userObject := user.NewUsers(mockClient, ens, logger)
		pass := "password1password1"
		_, mnemonic, _, _, ui, err := userObject.CreateNewUserV2(user1, pass, "", "", tm)
		if err != nil {
			t.Fatal(err)
		}

		_, _, _, _, _, err = userObject.CreateNewUserV2(user1, pass, "", "", tm)
		if !errors.Is(err, user.ErrUserAlreadyPresent) {
			t.Fatal(err)
		}

		// validate user
		if !userObject.IsUsernameAvailableV2(ui.GetUserName()) {
			t.Fatalf("user not created")
		}
		if !userObject.IsUserNameLoggedIn(ui.GetUserName()) {
			t.Fatalf("user not loggin in")
		}
		if ui == nil {
			t.Fatalf("invalid user info")
		}
		if ui.GetUserName() != user1 {
			t.Fatalf("invalid user name")
		}
		if ui.GetFeed() == nil || ui.GetAccount() == nil {
			t.Fatalf("invalid feed or account")
		}
		err = ui.GetAccount().GetWallet().IsValidMnemonic(mnemonic)
		if err != nil {
			t.Fatalf("invalid mnemonic")
		}

		_, _, _, _, _, err = userObject.CreateNewUserV2(user1, pass+pass, mnemonic, "", tm)
		if err != nil {
			t.Fatal(err)
		}
	})
}
