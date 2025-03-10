/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2018, 2019 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package provider

import (
	"github.com/IBM/ibmcloud-storage-volume-lib/lib/provider"
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/models"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

// maxRetryAttempt ...
var maxRetryAttempt = 10

// maxRetryGap ...
var maxRetryGap = 60

// retryGap ...
var retryGap = 10

var volumeIDPartsCount = 5

var skipErrorCodes = map[string]bool{
	"validation_invalid_name":          true,
	"volume_capacity_max":              true,
	"volume_id_invalid":                true,
	"volume_profile_iops_invalid":      true,
	"volume_capacity_zero_or_negative": true,
	"not_found":                        true,
	"internal_error":                   false,
	"invalid_route":                    false,
}

// retry ...
func retry(logger *zap.Logger, retryfunc func() error) error {
	var err error

	for i := 0; i < maxRetryAttempt; i++ {
		if i > 0 {
			time.Sleep(time.Duration(retryGap) * time.Second)
		}
		err = retryfunc()
		if err != nil {
			//Skip retry for the below type of Errors
			modelError, ok := err.(*models.Error)
			if !ok {
				continue
			}
			if skipRetry(modelError) {
				break
			}
			if i >= 1 {
				retryGap = 2 * retryGap
				if retryGap > maxRetryGap {
					retryGap = maxRetryGap
				}
			}
			if (i + 1) < maxRetryAttempt {
				logger.Info("Error while executing the function. Re-attempting execution ..", zap.Int("attempt..", i+2), zap.Int("retry-gap", retryGap), zap.Int("max-retry-Attempts", maxRetryGap), zap.Error(err))
			}
			continue
		}
		return err
	}
	return err
}

// skipRetry skip retry as per listed error codes
func skipRetry(err *models.Error) bool {
	for _, errorItem := range err.Errors {
		skipStatus, ok := skipErrorCodes[string(errorItem.Code)]
		if ok {
			return skipStatus
		}
	}
	return false
}

// ToInt ...
func ToInt(valueInInt string) int {
	value, err := strconv.Atoi(valueInInt)
	if err != nil {
		return 0
	}
	return value
}

// ToInt64 ...
func ToInt64(valueInInt string) int64 {
	value, err := strconv.ParseInt(valueInInt, 10, 64)
	if err != nil {
		return 0
	}
	return value
}

// FromProviderToLibVolume converting vpc provider volume type to generic lib volume type
func FromProviderToLibVolume(vpcVolume *models.Volume, logger *zap.Logger) (libVolume *provider.Volume) {
	logger.Debug("Entry of FromProviderToLibVolume method...")
	defer logger.Debug("Exit from FromProviderToLibVolume method...")

	if vpcVolume == nil {
		logger.Info("Volume details are empty")
		return
	}

	if vpcVolume.Zone == nil {
		logger.Info("Volume zone is empty")
		return
	}

	logger.Debug("Volume details of VPC client", zap.Reflect("models.Volume", vpcVolume))

	volumeCap := int(vpcVolume.Capacity)
	iops := strconv.Itoa(int(vpcVolume.Iops))
	var createdDate time.Time
	if vpcVolume.CreatedAt != nil {
		createdDate = *vpcVolume.CreatedAt
	}

	libVolume = &provider.Volume{
		VolumeID:     vpcVolume.ID,
		Provider:     VPC,
		Capacity:     &volumeCap,
		Iops:         &iops,
		VolumeType:   VolumeType,
		CreationTime: createdDate,
	}
	if vpcVolume.Zone != nil {
		libVolume.Region = vpcVolume.Zone.Name
	}
	return
}

// IsValidVolumeIDFormat validating
func IsValidVolumeIDFormat(volID string) bool {
	parts := strings.Split(volID, "-")
	if len(parts) != volumeIDPartsCount {
		return false
	}
	return true
}

// SetRetryParameters sets the retry logic parameters
func SetRetryParameters(maxAttempts int, maxGap int) {
	if maxAttempts > 0 {
		maxRetryAttempt = maxAttempts
	}

	if maxGap > 0 {
		maxRetryGap = maxGap
	}
}
