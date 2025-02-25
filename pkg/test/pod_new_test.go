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
	"os"
	"strings"
	"testing"
	"time"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

func TestPodNew(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(os.Stdout, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	pod1 := pod.NewPod(mockClient, fd, acc, tm, logger)

	podName1 := "test1"
	podName2 := "test2"
	t.Run("create-first-pod", func(t *testing.T) {
		podPresent := pod1.IsPodPresent("")
		if podPresent {
			t.Fatal("blank podname should not be present")
		}

		// check too long pod name
		randomLongPOdName, err := utils.GetRandString(65)
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		_, err = pod1.CreatePod(randomLongPOdName, "", podPassword)
		if !errors.Is(err, pod.ErrTooLongPodName) {
			t.Fatalf("error creating pod %s", podName1)
		}
		pod1Present := pod1.IsPodPresent(randomLongPOdName)
		if pod1Present {
			t.Fatal("pod1 should not be present")
		}
		info, err := pod1.CreatePod(podName1, "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s: %s", podName1, err.Error())
		}

		if pod1.GetFeed() == nil || pod1.GetAccount() == nil {
			t.Fatalf("userAddress not initialized")
		}

		if info.GetPodName() != podName1 {
			t.Fatalf("invalid pod name: expected %s got %s", podName1, info.GetPodName())
		}

		pods, _, err := pod1.ListPods()
		if err != nil {
			t.Fatalf("error getting pods")
		}

		if len(pods) != 1 {
			t.Fatalf("length of pods is not 1")
		}

		if strings.Trim(pods[0], "\n") != podName1 {
			t.Fatalf("podName is not %s", podName1)
		}

		infoGot, _, err := pod1.GetPodInfoFromPodMap(podName1)
		if err != nil {
			t.Fatalf("could not get pod from podMap")
		}

		if infoGot.GetPodName() != podName1 {
			t.Fatalf("invalid pod name: expected %s got %s", podName1, infoGot.GetPodName())
		}
	})

	t.Run("create-second-pod", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		info, err := pod1.CreatePod(podName2, "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s", podName2)
		}

		if info.GetPodName() != podName2 {
			t.Fatalf("invalid pod name: expected %s got %s", podName2, info.GetPodName())
		}

		pods, _, err := pod1.ListPods()
		if err != nil {
			t.Fatalf("error getting pods")
		}

		if len(pods) != 2 {
			t.Fatalf("length of pods is not 2")
		}

		if strings.Trim(pods[0], "\n") != podName2 && strings.Trim(pods[1], "\n") != podName2 {
			t.Fatalf("podName is not %s", podName2)
		}

		infoGot, _, err := pod1.GetPodInfoFromPodMap(podName2)
		if err != nil {
			t.Fatalf("could not get pod from podMap")
		}

		if infoGot.GetPodName() != podName2 {
			t.Fatalf("invalid pod name: expected %s got %s", podName2, infoGot.GetPodName())
		}
	})
}
