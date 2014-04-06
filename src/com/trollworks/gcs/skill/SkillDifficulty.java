/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.skill;

/** The possible skill difficulty levels. */
public enum SkillDifficulty {
	/** The "easy" difficulty. */
	E,
	/** The "average" difficulty. */
	A,
	/** The "hard" difficulty. */
	H,
	/** The "very hard" difficulty. */
	VH,
	/** The "wildcard" difficulty. */
	W {

		@Override
		public int getBaseRelativeLevel() {
			return VH.getBaseRelativeLevel();
		}
	};

	/** @return The base relative skill level at 0 points. */
	public int getBaseRelativeLevel() {
		return -ordinal();
	}
}
