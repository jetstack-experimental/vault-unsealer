package cmd

import (
	"fmt"
	"github.com/jetstack/vault-unsealer/pkg/kv/alicloud_kms"
	"github.com/jetstack/vault-unsealer/pkg/kv/alicloud_oss"
	"github.com/spf13/viper"
	"os"

	"github.com/jetstack/vault-unsealer/pkg/kv"
	"github.com/jetstack/vault-unsealer/pkg/kv/aws_kms"
	"github.com/jetstack/vault-unsealer/pkg/kv/aws_ssm"
	"github.com/jetstack/vault-unsealer/pkg/kv/cloudkms"
	"github.com/jetstack/vault-unsealer/pkg/kv/gcs"
	"github.com/jetstack/vault-unsealer/pkg/kv/local"
	"github.com/jetstack/vault-unsealer/pkg/vault"
)

func vaultConfigForConfig(cfg *viper.Viper) (vault.Config, error) {

	return vault.Config{
		KeyPrefix: "vault",

		SecretShares:    appConfig.GetInt(cfgSecretShares),
		SecretThreshold: appConfig.GetInt(cfgSecretThreshold),

		InitRootToken:  appConfig.GetString(cfgInitRootToken),
		StoreRootToken: appConfig.GetBool(cfgStoreRootToken),

		OverwriteExisting: appConfig.GetBool(cfgOverwriteExisting),
	}, nil
}

func kvStoreForConfig(cfg *viper.Viper) (kv.Service, error) {

	switch cfg.GetString(cfgMode) {
	case cfgModeValueGoogleCloudKMSGCS:

		g, err := gcs.New(
			cfg.GetString(cfgGoogleCloudStorageBucket),
			cfg.GetString(cfgGoogleCloudStoragePrefix),
		)

		if err != nil {
			return nil, fmt.Errorf("error creating google cloud storage kv store: %s", err.Error())
		}

		kms, err := cloudkms.New(g,
			cfg.GetString(cfgGoogleCloudKMSProject),
			cfg.GetString(cfgGoogleCloudKMSLocation),
			cfg.GetString(cfgGoogleCloudKMSKeyRing),
			cfg.GetString(cfgGoogleCloudKMSCryptoKey),
		)

		if err != nil {
			return nil, fmt.Errorf("error creating google cloud kms kv store: %s", err.Error())
		}

		return kms, nil

	case cfgModeValueAWSKMSSSM:
		ssm, err := aws_ssm.New(cfg.GetString(cfgAWSSSMKeyPrefix))
		if err != nil {
			return nil, fmt.Errorf("error creating AWS SSM kv store: %s", err.Error())
		}

		kms, err := aws_kms.New(ssm, cfg.GetString(cfgAWSKMSKeyID))
		if err != nil {
			return nil, fmt.Errorf("error creating AWS KMS ID kv store: %s", err.Error())
		}

		return kms, nil

	case cfgModeValueAlicloudKMSOSS:
		envAlicloudAccessKey := os.Getenv("ALICLOUD_ACCESS_KEY")
		envAlicloudSecretKey := os.Getenv("ALICLOUD_SECRET_KEY")

		o, err := alicloud_oss.New(
			cfg.GetString(cfgAlicloudStorageEndpoint),
			cfg.GetString(cfgAlicloudStorageBucket),
			cfg.GetString(cfgAlicloudStoragePrefix),
			envAlicloudSecretKey,
			envAlicloudSecretKey)

		if err != nil {
			return nil, fmt.Errorf("error creating Alicloud storage kv store: %s", err.Error())
		}


		kms, err := alicloud_kms.New(o,
			cfg.GetString(cfgAlicloudKMSKeyID),
			cfg.GetString(cfgAlicloudKMSRegion),
			envAlicloudAccessKey,
			envAlicloudSecretKey)

		if err != nil {
			return nil, fmt.Errorf("error creating Alicloud KMS ID kv store: %s", err.Error())
		}

		return kms, nil

	case cfgModeValueLocal:
		return local.New(cfg.GetString(cfgLocalKeyDir))

	default:
		return nil, fmt.Errorf("Unsupported backend mode: '%s'", cfg.GetString(cfgMode))
	}
}
