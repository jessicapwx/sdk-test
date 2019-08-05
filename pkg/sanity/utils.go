/*
Copyright 2018 Portworx

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

package sanity

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"
	auth "github.com/libopenstorage/sdk-test/pkg/auth"
	common "github.com/libopenstorage/sdk-test/pkg/common"
	"google.golang.org/grpc/metadata"

	. "github.com/onsi/gomega"
)

const (
	BYTE = 1.0 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
)

func createUsersTokens() map[string]string {

	users := make(map[string]string)

	// user1
	user1 := createToken(&auth.Claims{
		Subject: "user1",
		Name:    "user1",
		Email:   "user1@user",
		Roles:   []string{"system.user"},
		Groups:  []string{"users"},
	}, &auth.Options{
		Expiration: time.Now().Add(1 * time.Hour).Unix(),
	}, config.SharedSecret)
	users["user1"] = user1

	// user1
	user2 := createToken(&auth.Claims{
		Subject: "user2",
		Name:    "user2",
		Email:   "user2@user",
		Roles:   []string{"system.user"},
		Groups:  []string{"users"},
	}, &auth.Options{
		Expiration: time.Now().Add(1 * time.Hour).Unix(),
	}, config.SharedSecret)
	users["user2"] = user2

	// user1
	user3 := createToken(&auth.Claims{
		Subject: "user3",
		Name:    "user3",
		Email:   "user3@user",
		Roles:   []string{"system.user"},
		Groups:  []string{"users", "testers"},
	}, &auth.Options{
		Expiration: time.Now().Add(1 * time.Hour).Unix(),
	}, config.SharedSecret)
	users["user3"] = user3

	// admin
	admin := createToken(&auth.Claims{
		Subject: "admin",
		Name:    "admin",
		Email:   "admin@user",
		Roles:   []string{"system.admin"},
		Groups:  []string{"*"},
	}, &auth.Options{
		Expiration: time.Now().Add(1 * time.Hour).Unix(),
	}, config.SharedSecret)
	users["admin"] = admin

	// expired
	expired := createToken(&auth.Claims{
		Subject: "expired",
		Name:    "expired",
		Email:   "expired@user",
		Roles:   []string{"system.view"},
	}, &auth.Options{
		Expiration: time.Now().Add(-1 * time.Hour).Unix(),
	}, config.SharedSecret)
	users["expired"] = expired

	return users
}

func createXUsersTokens(prefix string, amount int) map[string]string {
	users := common.NewConcStrToStrMap()
	if amount <= 0 {
		return users.GetKeyValMap()
	}
	var wg sync.WaitGroup
	for i := 0; i < amount; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			user := createToken(&auth.Claims{
				Subject: name,
				Name:    name,
				Email:   fmt.Sprintf("%s@portworx.com", name),
				Roles:   []string{"system.user"},
				Groups:  []string{"users"},
			}, &auth.Options{
				Expiration: time.Now().Add(1 * time.Hour).Unix(),
			}, config.SharedSecret)
			users.Add(name, user)
		}(name)
	}
	wg.Wait()
	return users.GetKeyValMap()
}

func setContextWithToken(ctx context.Context, token string) context.Context {
	md := metadata.New(map[string]string{
		"authorization": "bearer " + token,
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func testVolumeDetails(
	req *api.SdkVolumeCreateRequest,
	volume *api.Volume,
) {

	// check volume specs
	Expect(volume.Spec.Ephemeral).To(BeEquivalentTo(req.Spec.Ephemeral))
	// Let's skip testing the block size
	//Expect(volume.Spec.BlockSize).To(BeEquivalentTo(req.Spec.BlockSize))
	Expect(volume.Spec.Cascaded).To(BeEquivalentTo(req.Spec.Cascaded))
	Expect(volume.Spec.Compressed).To(BeEquivalentTo(req.Spec.Compressed))

	Expect(volume.Spec.Dedupe).To(BeEquivalentTo(req.Spec.Dedupe))

	Expect(volume.Spec.Group).To(BeEquivalentTo(req.Spec.Group))
	Expect(volume.Spec.GroupEnforced).To(BeEquivalentTo(req.Spec.GroupEnforced))

	Expect(volume.Spec.Journal).To(BeEquivalentTo(req.Spec.Journal))
	Expect(volume.Spec.Sharedv4).To(BeEquivalentTo(req.Spec.Sharedv4))
	Expect(volume.Spec.Passphrase).To(BeEquivalentTo(req.Spec.Passphrase))
	Expect(volume.Spec.ReplicaSet).To(BeEquivalentTo(req.Spec.ReplicaSet))
	Expect(volume.Spec.Scale).To(BeEquivalentTo(req.Spec.Scale))
	Expect(volume.Spec.Shared).To(BeEquivalentTo(req.Spec.Shared))
	Expect(volume.Spec.Size).To(BeEquivalentTo(req.Spec.Size))
	Expect(volume.Spec.SnapshotInterval).To(BeEquivalentTo(req.Spec.SnapshotInterval))
	Expect(volume.Spec.SnapshotSchedule).To(BeEquivalentTo(req.Spec.SnapshotSchedule))
	Expect(volume.Source.Parent).To(BeEmpty())
	Expect(volume.Locator.Name).To(BeEquivalentTo(req.Name))

	// TODO: Fake driver mmust honor the below parameters

	//Expect(volume.Spec.AggregationLevel).To(BeEquivalentTo(req.Spec.AggregationLevel))
	//Expect(volume.Spec.Cos).To(BeEquivalentTo(req.Spec.Cos))
	//Expect(volume.Spec.Encrypted).To(BeEquivalentTo(req.Spec.Encrypted))
	//Expect(volume.Spec.Format).To(BeEquivalentTo(req.Spec.Format))
	//Expect(volume.Spec.HaLevel).To(BeEquivalentTo(req.Spec.HaLevel))
	//Expect(volume.Spec.IoProfile).To(BeEquivalentTo(req.Spec.IoProfile))
	//Expect(volume.Spec.Sticky).To(BeEquivalentTo(req.Spec.Sticky))
}

func testVolumeCreation(req *api.SdkVolumeCreateRequest) {

}

// This will create credential for provider listed from cb.yaml file
func parseAndCreateCredentials(credClient api.OpenStorageCredentialsClient) int {
	numCredCreated := 0
	for provider, providerParams := range config.ProviderConfig.CloudProviders {
		if provider == "aws" {
			credReq := &api.SdkCredentialCreateRequest{
				Name: providerParams["CredName"],
				CredentialType: &api.SdkCredentialCreateRequest_AwsCredential{
					AwsCredential: &api.SdkAwsCredentialRequest{
						AccessKey:  providerParams["CredAccessKey"],
						SecretKey:  providerParams["CredSecretKey"],
						Endpoint:   providerParams["CredEndpoint"],
						Region:     providerParams["CredRegion"],
						DisableSsl: providerParams["CredDisableSSL"] == "true",
					},
				},
			}

			credResp, err := credClient.Create(setContextWithToken(context.Background(), users["admin"]), credReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(credResp.GetCredentialId()).NotTo(BeEmpty())
			numCredCreated++

		} else if provider == "azure" {
			credReq := &api.SdkCredentialCreateRequest{
				Name: providerParams["CredName"],
				CredentialType: &api.SdkCredentialCreateRequest_AzureCredential{
					AzureCredential: &api.SdkAzureCredentialRequest{
						AccountKey:  providerParams["CredAccountName"],
						AccountName: providerParams["CredAccountKey"],
					},
				},
			}

			credResp, err := credClient.Create(setContextWithToken(context.Background(), users["admin"]), credReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(credResp.GetCredentialId()).NotTo(BeEmpty())
			numCredCreated++

		} else if provider == "google" {
			credReq := &api.SdkCredentialCreateRequest{
				Name: providerParams["CredName"],
				CredentialType: &api.SdkCredentialCreateRequest_GoogleCredential{
					GoogleCredential: &api.SdkGoogleCredentialRequest{
						ProjectId: providerParams["CredProjectID"],
						JsonKey:   providerParams["CredJsonKey"],
					},
				},
			}

			credResp, err := credClient.Create(setContextWithToken(context.Background(), users["admin"]), credReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(credResp.GetCredentialId()).NotTo(BeEmpty())
			numCredCreated++
		}
	}
	return numCredCreated
}

func newTestVolume(volClient api.OpenStorageVolumeClient) string {
	volReq := &api.SdkVolumeCreateRequest{
		Name: fmt.Sprintf("sdk-vol-%v", time.Now().Unix()),
		Spec: &api.VolumeSpec{
			Size:      uint64(5 * GIGABYTE),
			Shared:    false,
			HaLevel:   3,
			IoProfile: api.IoProfile_IO_PROFILE_DB,
			Cos:       api.CosType_HIGH,
			Format:    api.FSType_FS_TYPE_XFS,
		},
	}
	volResp, err := volClient.Create(setContextWithToken(context.Background(), users["admin"]), volReq)
	Expect(err).NotTo(HaveOccurred())
	Expect(volResp).NotTo(BeNil())
	Expect(volResp.VolumeId).NotTo(BeEmpty())
	volID := volResp.VolumeId
	return volID
}

func newTestCredential(credClient api.OpenStorageCredentialsClient) string {
	credReq := &api.SdkCredentialCreateRequest{
		Name: "test-credential",
		CredentialType: &api.SdkCredentialCreateRequest_AwsCredential{
			AwsCredential: &api.SdkAwsCredentialRequest{
				AccessKey: "aws-access-key",
				SecretKey: "AWS_SECRET_KEY_$$",
				Endpoint:  "s3.aws.com",
				Region:    "us-east",
			},
		},
	}

	credResp, err := credClient.Create(setContextWithToken(context.Background(), users["admin"]), credReq)
	Expect(err).NotTo(HaveOccurred())
	Expect(credResp.GetCredentialId()).NotTo(BeEmpty())
	return credResp.GetCredentialId()
}

// This will create credential for provider listed from cb.yaml file
func parseAndCreateCredentials2(credClient api.OpenStorageCredentialsClient) map[string]string {
	credMap := make(map[string]string)
	for provider, providerParams := range config.ProviderConfig.CloudProviders {
		if provider == "aws" {

			credReq := &api.SdkCredentialCreateRequest{
				Name: providerParams["CredName"],
				CredentialType: &api.SdkCredentialCreateRequest_AwsCredential{
					AwsCredential: &api.SdkAwsCredentialRequest{
						AccessKey:  providerParams["CredAccessKey"],
						SecretKey:  providerParams["CredSecretKey"],
						Endpoint:   providerParams["CredEndpoint"],
						Region:     providerParams["CredRegion"],
						DisableSsl: providerParams["CredDisableSSL"] == "true",
					},
				},
			}

			credResp, err := credClient.Create(setContextWithToken(context.Background(), users["admin"]), credReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(credResp.GetCredentialId()).NotTo(BeEmpty())
			credMap["aws"] = credResp.GetCredentialId()

		} else if provider == "azure" {
			credReq := &api.SdkCredentialCreateRequest{
				Name: providerParams["CredName"],
				CredentialType: &api.SdkCredentialCreateRequest_AzureCredential{
					AzureCredential: &api.SdkAzureCredentialRequest{
						AccountKey:  providerParams["CredAccountName"],
						AccountName: providerParams["CredAccountKey"],
					},
				},
			}
			credResp, err := credClient.Create(setContextWithToken(context.Background(), users["admin"]), credReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(credResp.GetCredentialId()).NotTo(BeEmpty())
			credMap["azure"] = credResp.GetCredentialId()

		} else if provider == "google" {
			credReq := &api.SdkCredentialCreateRequest{
				Name: providerParams["CredName"],
				CredentialType: &api.SdkCredentialCreateRequest_GoogleCredential{
					GoogleCredential: &api.SdkGoogleCredentialRequest{
						ProjectId: providerParams["CredProjectID"],
						JsonKey:   providerParams["CredJsonKey"],
					},
				},
			}

			credResp, err := credClient.Create(setContextWithToken(context.Background(), users["admin"]), credReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(credResp.GetCredentialId()).NotTo(BeEmpty())
			credMap["google"] = credResp.GetCredentialId()
		}
	}
	return credMap
}

func isCapabilitySupported(c api.OpenStorageIdentityClient,
	capType api.SdkServiceCapability_OpenStorageService_Type,
) bool {

	caps, err := c.Capabilities(
		setContextWithToken(context.Background(), users["admin"]),
		&api.SdkIdentityCapabilitiesRequest{})
	Expect(err).NotTo(HaveOccurred())
	Expect(caps).NotTo(BeNil())
	Expect(caps.GetCapabilities()).NotTo(BeNil())

	for _, cap := range caps.GetCapabilities() {
		Expect(cap.GetService()).NotTo(BeNil())
		if cap.GetService().GetType() == capType {
			return true
		}
	}
	return false
}

func createToken(claims *auth.Claims, options *auth.Options, sharedSecret string) string {

	if len(claims.Issuer) == 0 {
		claims.Issuer = config.Issuer
	}

	// This never fails
	signature, _ := auth.NewSignatureSharedSecret(sharedSecret)

	token, err := auth.Token(claims, signature, options)
	Expect(err).NotTo(HaveOccurred())

	return token
}

// f() return <wait as bool, or err>.
// waitFor stops when f() return (false) or err != nil
func waitFor(timeout time.Duration, period time.Duration, f func() (bool, error)) error {
	timeoutChan := time.After(timeout)
	var (
		wait bool = true
		err  error
	)
	for wait {
		select {
		case <-timeoutChan:
			return fmt.Errorf("Timed out")
		default:
			wait, err = f()
			if err != nil {
				return err
			}
			time.Sleep(period)
		}
	}

	return nil
}

func deleteVol(ctx context.Context, vc api.OpenStorageVolumeClient, volid string) error {
	err := waitFor(5*time.Minute, time.Second, func() (bool, error) {
		respInspectVol, err := vc.Inspect(
			ctx,
			&api.SdkVolumeInspectRequest{
				VolumeId: volid,
			})
		if err != nil {
			return false, err
		}
		if len(respInspectVol.GetVolume().GetAttachedOn()) == 0 {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return err
	}

	_, err = vc.Delete(
		ctx,
		&api.SdkVolumeDeleteRequest{
			VolumeId: volid,
		})
	return err
}

func printVolumeDetails(
	volume *api.Volume,
) {
	fmt.Printf("volume ID: %s\n", volume.Id)
}

func summarizeErrorsFromStringErrorChanMap(stringErrMap map[string]chan (error)) error {
	var summarizedErrMsg string
	for key, errChan := range stringErrMap {
		err := <-errChan
		if err != nil {
			summarizedErrMsg = fmt.Sprintf("%s\nerror of %s: %v", summarizedErrMsg, key, err)
		}
	}
	fmt.Printf("\nsummarizedErrMsg: %s", summarizedErrMsg)
	if strings.TrimSpace(summarizedErrMsg) == "" {
		return nil
	}
	return fmt.Errorf("%s", summarizedErrMsg)
}
