// Copyright 2021 Upbound Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package aws contains logic to handle AWS Marketplace billing.
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/marketplacemetering"
	"github.com/golang-jwt/jwt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// These constants are given by AWS Marketplace.
// TODO(muvaf): Consider fetching them from an Upbound API but keep the latest
// ones hard-coded as fallback for air-gapped environments.
const (
	MarketplaceProductCode      = "1fszvu527waovqeuhpkyx2b5d"
	MarketplacePublicKey        = "-----BEGIN PUBLIC KEY-----\nMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAyu7Xq7XTBRgFWCL+DXj8\nXyc/fPLWNQ1adPDf8zqkJ1H1JCTg6fUo7HUvNu0BAbPwIME4aDEzteJkhPq9IzS8\nHlrZT/7DqSPV9bXnR9OkqugfbFPyHGyd9afHyfDJfGwfqBP5r8oBuGwmCw5Ia088\nAcePfkVEisAo+8KiBAE16bqvDw0v5YzDrDVpHH9YdK1q9eG5WRTt0h7lYFj8dydr\nh+OyONGyWTkAWbs3JpsQLZgRdU6Klj5aZzO6FeUc2kOz2Hs+QvKgbNSpgV0000KK\n2on4L1+WJau7sj8EFquFdk2C0MhucIy6ceWXGB3YAOb8c0H9FT0eSY5rtX154otW\njmV9vMLLX1gajtQD0iOLBLRQ3WliP7fGc6o3StjMrbKh+ErXGVzzJnjK2eQhgkg/\n/DgcKjUptZ21gdbqbQBGwvfitBEJX7VCwF4VMhFM8JQiAxCVBZ7kkY5ZlGjvN2gO\nAMFKarvAWRwrZisxKWe+RFBU1EI5WS75X7owU/IehIabAgMBAAE=\n-----END PUBLIC KEY-----\n"
	MarketplacePublicKeyVersion = 1
)

// SecretKeyAWSMeteringSignature is the key whose value contains JWT signature returned
// from AWS Metering Service.
const (
	SecretKeyAWSMeteringSignature = "awsMeteringSignature"

	errRegisterUsage            = "cannot register usage"
	errApplySecret              = "cannot apply entitlement secret"
	errParseToken               = "cannot parse token"
	errProductCodeMatchFmt      = "productCode %s does not match expected %s"
	errNonceMatchFmt            = "nonce %s does not match expected %s"
	errPublicKeyVersionMatchFmt = "publicKeyVersion %s does not match expected %f"
)

type marketplaceClient interface {
	RegisterUsage(ctx context.Context, params *marketplacemetering.RegisterUsageInput, optFns ...func(*marketplacemetering.Options)) (*marketplacemetering.RegisterUsageOutput, error)
}

// NewMarketplace returns a new Marketplace object that can register usage.
func NewMarketplace(cl client.Client, mcl marketplaceClient, publicKey string) *Marketplace {
	return &Marketplace{
		client:    resource.NewApplicatorWithRetry(resource.NewAPIPatchingApplicator(cl), resource.IsAPIErrorWrapped, &retry.DefaultRetry),
		metering:  mcl,
		publicKey: publicKey,
	}
}

// Marketplace implements Registerer for AWS Marketplace API.
type Marketplace struct {
	client    resource.Applicator
	metering  marketplaceClient
	publicKey string
}

// Register makes sure user is entitled for this usage in an idempotent way.
func (am *Marketplace) Register(ctx context.Context, s *v1.Secret, uid string) (string, error) {
	if len(s.Data[SecretKeyAWSMeteringSignature]) > 0 {
		return string(s.Data[SecretKeyAWSMeteringSignature]), nil
	}
	u := &marketplacemetering.RegisterUsageInput{
		ProductCode:      aws.String(MarketplaceProductCode),
		PublicKeyVersion: aws.Int32(MarketplacePublicKeyVersion),
		Nonce:            aws.String(uid),
	}
	resp, err := am.metering.RegisterUsage(ctx, u)
	if err != nil {
		return "", errors.Wrap(err, errRegisterUsage)
	}
	if s.Data == nil {
		s.Data = map[string][]byte{}
	}
	s.Data[SecretKeyAWSMeteringSignature] = []byte(aws.ToString(resp.Signature))
	return aws.ToString(resp.Signature), errors.Wrapf(am.client.Apply(ctx, s), errApplySecret)
}

// Verify makes sure the signature is signed by AWS Marketplace.
func (am *Marketplace) Verify(token, uid string) (bool, error) {
	t, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(_ *jwt.Token) (any, error) {
		return jwt.ParseRSAPublicKeyFromPEM([]byte(am.publicKey))
	})
	if err != nil {
		return false, errors.Wrap(err, errParseToken)
	}
	if !t.Valid {
		return false, nil
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return false, errors.Errorf("expected jwt.MapClaims, got %t instead", t.Claims)
	}
	switch {
	case claims["productCode"] != MarketplaceProductCode:
		return false, errors.Errorf(errProductCodeMatchFmt, claims["productCode"], MarketplaceProductCode)
	case claims["nonce"] != uid:
		return false, errors.Errorf(errNonceMatchFmt, claims["nonce"], uid)
	case claims["publicKeyVersion"] != float64(MarketplacePublicKeyVersion):
		return false, errors.Errorf(errPublicKeyVersionMatchFmt, claims["publicKeyVersion"], float64(MarketplacePublicKeyVersion))
	}
	return true, nil
}
