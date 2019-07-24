package sanity

import (
	"context"
	"fmt"
	"sync"
	"time"

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"
	common "github.com/libopenstorage/sdk-test/pkg/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var scaleUsers map[string]string
var userVolumeMap *common.ConcMap

type VolumeRequest struct {
	VolID         string
	CreateRequest *api.SdkVolumeCreateRequest
	Token         string
}

var _ = Describe("Security Scale", func() {
	var (
		c  api.OpenStorageVolumeClient
		ic api.OpenStorageIdentityClient
	)

	BeforeEach(func() {
		c = api.NewOpenStorageVolumeClient(conn)
		ic = api.NewOpenStorageIdentityClient(conn)
		isSupported := isCapabilitySupported(
			ic,
			api.SdkServiceCapability_OpenStorageService_VOLUME,
		)
		if !isSupported {
			Fail("Volume capability not supported , skipping related tests")
		}
	})

	AfterEach(func() {
	})

	Describe("Security", func() {

		BeforeEach(func() {
		})

		It("Should be able to create the users", func() {
			By("Creating users")
			scaleUsers = createXUsersTokens("scaleUsers", 30)
		})
		It("Should be able to create the volumes", func() {
			By("Creating volumes")
			err := createVolumesConcurrently(c)
			Expect(err).NotTo(HaveOccurred())
		})
		It("Should be able to inspect the created Volume", func() {
			By("Inspecting volumes")
			err := inspectVolumesConcurrently(c)
			Expect(err).NotTo(HaveOccurred())
		})
		It("Owner Should be able to delete its own Volume", func() {
			By("Deleting volumes")
			err := deleteVolumesConcurrently(c)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})

func createVolumesConcurrently(c api.OpenStorageVolumeClient) error {
	//userVolumeMap is mapping user'name to volumes' ID
	userVolumeMap = common.NewConcMap()
	var volErrorMap = common.NewConcStringErrChanMap()
	var wg sync.WaitGroup
	for name, userToken := range scaleUsers {
		userName := name
		token := userToken
		wg.Add(1)
		go func(userName string, token string) {
			defer wg.Done()
			t := time.Now()
			tstr := t.Format("20060102150405")
			req := &api.SdkVolumeCreateRequest{
				Name: "sdk-vol-" + tstr + "-" + userName,
				Spec: &api.VolumeSpec{
					Size:    uint64(5 * GIGABYTE),
					HaLevel: 2,
				},
			}
			createResponse, err := c.Create(setContextWithToken(context.Background(), token), req)
			volID := createResponse.VolumeId
			resp, err := c.Inspect(
				setContextWithToken(context.Background(), token),
				&api.SdkVolumeInspectRequest{
					VolumeId: volID,
				},
			)
			printVolumeDetails(resp.Volume)
			userVolumeMap.Add(userName, resp.GetVolume().GetId())
			errChan := make(chan (error), 1)
			errChan <- err
			volErrorMap.Add(volID, errChan)
		}(userName, token)
	}
	wg.Wait()
	for user, volID := range userVolumeMap.GetKeyValMap() {
		fmt.Printf("\nuser %s createdvolume ->: %s", user, volID)
	}
	return summarizeErrorsFromStringErrorChanMap(volErrorMap.GetKeyValMap())
}

func inspectVolumesConcurrently(c api.OpenStorageVolumeClient) error {
	var wg sync.WaitGroup
	var volErrorMap = common.NewConcStringErrChanMap()
	for user, id := range userVolumeMap.GetKeyValMap() {
		userName := user.(string)
		volID := id.(string)
		fmt.Printf("\nNow user %s is going to inspect volume %s", userName, volID)
		token := scaleUsers[userName]
		wg.Add(1)
		go func(userName string, volID string, token string) {
			defer wg.Done()
			_, err := c.Inspect(
				setContextWithToken(context.Background(), token),
				&api.SdkVolumeInspectRequest{
					VolumeId: volID,
				},
			)
			errChan := make(chan (error), 1)
			errChan <- err
			volErrorMap.Add(volID, errChan)
		}(userName, volID, token)
	}
	wg.Wait()
	//Receiving all errors from channels
	return summarizeErrorsFromStringErrorChanMap(volErrorMap.GetKeyValMap())
}

func deleteVolumesConcurrently(c api.OpenStorageVolumeClient) error {
	var wg sync.WaitGroup
	var volErrorMap = common.NewConcStringErrChanMap()
	for user, id := range userVolumeMap.GetKeyValMap() {
		userName := user.(string)
		volID := id.(string)
		token := scaleUsers[userName]
		fmt.Printf("\nNow user %s is going to delete volume %s", userName, volID)
		wg.Add(1)
		go func(userName string, volID string, token string) {
			defer wg.Done()
			err := deleteVol(
				setContextWithToken(context.Background(), token),
				c,
				volID,
			)
			errChan := make(chan (error), 1)
			errChan <- err
			volErrorMap.Add(volID, errChan)
		}(userName, volID, token)
	}
	wg.Wait()
	return summarizeErrorsFromStringErrorChanMap(volErrorMap.GetKeyValMap())
}
