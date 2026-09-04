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
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Venafi/vcert/v5/pkg/playbook/app/domain"
	"github.com/Venafi/vcert/v5/pkg/playbook/app/installer"
	"github.com/Venafi/vcert/v5/pkg/playbook/app/vcertutil"
)

// pickupFirstAttempt implements pickup-first mode (request.pickupFirst=true).
// Supported backends: TPP and NGTS. On other backends (Firefly, etc.)
// the feature is a silent no-op and the standard playbook flow runs.
//
// Decision flow on each run:
//
//   1. Locate the platform's "current" cert for this CN. TPP returns its
//      cert-object DN + metadata; NGTS queries certificate search by CN
//      and picks the latest valid certificate.
//   2. Cheap compare (thumbprint + ValidTo) against the installed cert:
//        match                -> defer to existing renewBefore check (handled=false).
//        platform older       -> refuse downgrade (no action, handled=true, errs=nil).
//        platform newer
//        or nothing installed -> proceed to step 3.
//   3. Full pickup of cert+chain (+ key if service-generated/vaulted) using the
//      platform-appropriate id; install at the playbook's paths via
//      the existing installer chain.
//   4. If anything goes wrong (locator unsupported, cert missing,
//      pickup error) return handled=false so the caller falls through
//      to the existing enroll flow.
func pickupFirstAttempt(config domain.Config, task domain.CertificateTask) (handled bool, errs []error) {
	if !task.Request.PickupFirst || config.ForceRenew {
		return false, nil
	}

	loc, err := vcertutil.LocateLatestCN(config, task.Request)
	if err != nil {
		if err == vcertutil.ErrLocateNotSupported {
			zap.L().Info("pickupFirst: not supported on this platform; running standard playbook flow",
				zap.String("platform", config.Connection.GetConnectorType().String()))
			return false, nil
		}
		zap.L().Info("pickupFirst: locator failed; falling through to enroll", zap.Error(err))
		return false, nil
	}
	if loc == nil || !loc.Found {
		zap.L().Info("pickupFirst: no matching cert on platform; falling through to enroll")
		return false, nil
	}

	installedThumb, installedNotAfter, foundInstalled := firstInstalledCertInfo(task.Installations)
	zap.L().Info("pickupFirst: located platform cert",
		zap.String("platform.thumbprint", loc.Thumbprint),
		zap.Time("platform.validTo", loc.ValidTo),
		zap.Bool("installed.found", foundInstalled),
		zap.String("installed.thumbprint", installedThumb),
	)

	normInstalled := vcertutil.NormalizeThumbprint(installedThumb)
	normPlatform := vcertutil.NormalizeThumbprint(loc.Thumbprint)

	if foundInstalled && normInstalled != "" && normInstalled == normPlatform {
		zap.L().Info("pickupFirst: thumbprint matches installed; deferring to renewBefore check (fast)")
		return false, nil
	}
	if foundInstalled && loc.ValidTo.Before(installedNotAfter) {
		zap.L().Warn("pickupFirst: platform cert is OLDER than installed; refusing downgrade",
			zap.Time("platform.validTo", loc.ValidTo),
			zap.Time("installed.notAfter", installedNotAfter),
		)
		return true, nil
	}

	keyPassword := vcertutil.GeneratePassword()
	pcc, certReq, err := vcertutil.PickupCertificateByLocator(config, task.Request, loc, keyPassword, true)
	if err != nil {
		zap.L().Warn("pickupFirst: pickup with key failed", zap.Error(err))
		pcc, certReq, err = vcertutil.PickupCertificateByLocator(config, task.Request, loc, "", false)
		if err != nil {
			zap.L().Info("pickupFirst: cert-only pickup also failed; falling through to enroll", zap.Error(err))
			return false, nil
		}
		zap.L().Info("pickupFirst: cert-only pickup OK (no vaulted key available)")
	} else {
		zap.L().Info("pickupFirst: pickup with key OK")
	}
	if pcc == nil || pcc.Certificate == "" {
		return false, nil
	}

	x509Certificate, preparedPcc, err := installer.CreateX509Cert(pcc, certReq, true)
	if err != nil {
		zap.L().Warn("pickupFirst: could not prepare pickup result; falling through", zap.Error(err))
		return false, nil
	}

	zap.L().Info("pickupFirst: installing pickup result without enrollment")
	if task.SetEnvVars != nil {
		zap.L().Debug("setting environment variables")
		setEnvVars(task, x509Certificate, preparedPcc)
	}
	errs = make([]error, 0)
	for _, inst := range task.Installations {
		if e := runInstaller(inst, preparedPcc); e != nil {
			errs = append(errs, e)
		}
	}
	return true, errs
}

// firstInstalledCertInfo returns the SHA-1 thumbprint (uppercase hex,
// matching the format both TPP and NGTS use) and NotAfter of the first
// existing installed cert file in the task. ok=false if no installation
// cert file is present on disk.
func firstInstalledCertInfo(installations []domain.Installation) (thumbprint string, notAfter time.Time, ok bool) {
	for _, inst := range installations {
		if inst.File == "" {
			continue
		}
		cert, err := installer.LoadInstalledPEM(inst.File)
		if err != nil || cert == nil {
			continue
		}
		sum := sha1.Sum(cert.Raw)
		return strings.ToUpper(hex.EncodeToString(sum[:])), cert.NotAfter, true
	}
	return "", time.Time{}, false
}
