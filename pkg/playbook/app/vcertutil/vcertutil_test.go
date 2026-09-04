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

package vcertutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Venafi/vcert/v5/pkg/playbook/app/domain"
	"github.com/Venafi/vcert/v5/pkg/venafi"
	"github.com/Venafi/vcert/v5/pkg/venafi/ngts"
)

func TestFindNewestNGTSCert(t *testing.T) {
	t1 := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	t2 := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	t3 := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)

	certs := []ngts.Certificate{
		{
			Id:                "cert-1",
			Fingerprint:       "FP1",
			ValidityEnd:       t1,
			CertificateStatus: "ACTIVE",
		},
		{
			Id:                "cert-2",
			Fingerprint:       "FP2",
			ValidityEnd:       t2,
			CertificateStatus: "ACTIVE",
		},
		{
			Id:                "cert-retired",
			Fingerprint:       "FP3",
			ValidityEnd:       t3,
			CertificateStatus: "RETIRED",
		},
	}

	// Should pick cert-2 because cert-retired is retired even though t3 > t2
	best, bestEnd := findNewestNGTSCert(certs)
	assert.NotNil(t, best)
	assert.Equal(t, "cert-2", best.Id)
	assert.Equal(t, "FP2", best.Fingerprint)
	parsedT2, _ := time.Parse(time.RFC3339, t2)
	assert.Equal(t, parsedT2.Unix(), bestEnd.Unix())

	// If only retired certs exist, should pick newest overall
	onlyRetired := []ngts.Certificate{
		{
			Id:                "cert-r1",
			ValidityEnd:       t1,
			CertificateStatus: "RETIRED",
		},
		{
			Id:                "cert-r2",
			ValidityEnd:       t3,
			CertificateStatus: "RETIRED",
		},
	}
	bestR, bestREnd := findNewestNGTSCert(onlyRetired)
	assert.NotNil(t, bestR)
	assert.Equal(t, "cert-r2", bestR.Id)
	parsedT3, _ := time.Parse(time.RFC3339, t3)
	assert.Equal(t, parsedT3.Unix(), bestREnd.Unix())
}

func TestLocateLatestCN_UnsupportedPlatform(t *testing.T) {
	config := domain.Config{
		Connection: domain.Connection{
			Platform: venafi.Firefly,
		},
	}
	req := domain.PlaybookRequest{
		PickupFirst: true,
	}

	loc, err := LocateLatestCN(config, req)
	assert.Nil(t, loc)
	assert.Equal(t, ErrLocateNotSupported, err)
}
