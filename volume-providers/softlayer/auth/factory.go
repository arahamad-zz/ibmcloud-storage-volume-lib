/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package auth

import (
	"github.com/arahamad/ibmcloud-storage-volume-lib/config"
	"github.com/arahamad/ibmcloud-storage-volume-lib/provider/auth"
	"github.com/arahamad/ibmcloud-storage-volume-lib/provider/local"
	"github.com/arahamad/ibmcloud-storage-volume-lib/volume-providers/iam"
)

// contextCredentialsFactory ...
type contextCredentialsFactory struct {
	softlayerConfig      *config.SoftlayerConfig
	tokenExchangeService auth.TokenExchangeService
}

var _ local.ContextCredentialsFactory = &contextCredentialsFactory{}

// NewContextCredentialsFactory ...
func NewContextCredentialsFactory(bluemixConfig *config.BluemixConfig, softlayerConfig *config.SoftlayerConfig) (*contextCredentialsFactory, error) {
	tokenExchangeService, err := iam.NewTokenExchangeService(bluemixConfig)
	if err != nil {
		return nil, err
	}

	return &contextCredentialsFactory{
		softlayerConfig:      softlayerConfig,
		tokenExchangeService: tokenExchangeService,
	}, nil
}
