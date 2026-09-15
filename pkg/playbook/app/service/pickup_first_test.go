/*
 * Copyright Venafi, Inc. and CyberArk Software Ltd. ("CyberArk")
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Venafi/vcert/v5/pkg/playbook/app/domain"
	"github.com/Venafi/vcert/v5/pkg/playbook/app/vcertutil"
)

func TestNormalizeThumbprint(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"AA:BB:CC:DD", "AABBCCDD"},
		{"aa:bb:cc:dd", "AABBCCDD"},
		{"AA BB CC DD", "AABBCCDD"},
		{"aa.bb.cc.dd", "AABBCCDD"},
		{"AABBCCDD", "AABBCCDD"},
		{"", ""},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, vcertutil.NormalizeThumbprint(tc.input))
		})
	}
}

func TestPickupFirst_Disabled(t *testing.T) {
	config := domain.Config{}
	task := domain.CertificateTask{
		Request: domain.PlaybookRequest{
			PickupFirst: false,
		},
	}
	handled, errs := pickupFirstAttempt(config, task)
	assert.False(t, handled)
	assert.Nil(t, errs)
}

func TestPickupFirst_ForceRenew(t *testing.T) {
	config := domain.Config{ForceRenew: true}
	task := domain.CertificateTask{
		Request: domain.PlaybookRequest{
			PickupFirst: true,
		},
	}
	handled, errs := pickupFirstAttempt(config, task)
	assert.False(t, handled)
	assert.Nil(t, errs)
}

func TestFirstInstalledCertInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pickup-first-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Generate self-signed certificate
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	notBefore := time.Now().Add(-1 * time.Hour)
	notAfter := time.Now().Add(24 * time.Hour).Truncate(time.Second)

	template := x509.Certificate{
		SerialNumber: big.NewInt(12345),
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
	}

	certDer, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	certFile := filepath.Join(tmpDir, "cert.pem")
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer})
	err = os.WriteFile(certFile, certPem, 0600)
	require.NoError(t, err)

	sum := sha1.Sum(certDer)
	expectedThumb := strings.ToUpper(hex.EncodeToString(sum[:]))

	installations := []domain.Installation{
		{
			Type: domain.FormatPEM,
			File: certFile,
		},
	}

	thumbprint, parsedNotAfter, ok := firstInstalledCertInfo(installations)
	assert.True(t, ok)
	assert.Equal(t, expectedThumb, thumbprint)
	assert.Equal(t, notAfter.Unix(), parsedNotAfter.Unix())
}

func TestFirstInstalledCertInfo_MissingFile(t *testing.T) {
	installations := []domain.Installation{
		{
			Type: domain.FormatPEM,
			File: "C:\\nonexistent\\path\\cert.pem",
		},
	}
	thumbprint, _, ok := firstInstalledCertInfo(installations)
	assert.False(t, ok)
	assert.Empty(t, thumbprint)
}

func TestKeyMatchesCert(t *testing.T) {
	// Generate key 1 and cert 1
	key1, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	key1Der := x509.MarshalPKCS1PrivateKey(key1)
	key1Pem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: key1Der})

	// Generate key 2 (different key)
	key2, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	key2Der := x509.MarshalPKCS1PrivateKey(key2)
	key2Pem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: key2Der})

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "app.example.com",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),
	}

	certDer, err := x509.CreateCertificate(rand.Reader, &template, &template, &key1.PublicKey, key1)
	require.NoError(t, err)
	certPem := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer}))

	// 1. Matching key should return true
	assert.True(t, keyMatchesCert(certPem, key1Pem, ""))

	// 2. Mismatched key (different key) should return false
	assert.False(t, keyMatchesCert(certPem, key2Pem, ""))

	// 3. Empty inputs should return false
	assert.False(t, keyMatchesCert("", key1Pem, ""))
	assert.False(t, keyMatchesCert(certPem, nil, ""))
	assert.False(t, keyMatchesCert(certPem, []byte("invalid-key-data"), ""))
}
