/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package model

// AllFeatureTypesWithoutContainedWeightType holds the possible FeatureType values, minus the ContainedWeightReductionFeatureType.
var AllFeatureTypesWithoutContainedWeightType []FeatureType

func init() {
	for _, one := range AllFeatureType {
		if one != ContainedWeightReductionFeatureType {
			AllFeatureTypesWithoutContainedWeightType = append(AllFeatureTypesWithoutContainedWeightType, one)
		}
	}
}