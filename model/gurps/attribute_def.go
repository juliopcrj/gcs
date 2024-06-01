// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"bytes"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/eval"
	"github.com/richardwilkes/toolbox/xmath/crc"
)

// ReservedIDs holds a list of IDs that are reserved for internal use.
var ReservedIDs = []string{SkillID, ParryID, BlockID, SizeModifierID, "10"}

// AttributeDef holds the definition of an attribute.
type AttributeDef struct {
	AttributeDefData
	Order     int
	KeyPrefix string
}

// AttributeDefData holds the data that will be serialized for the AttributeDef.
type AttributeDefData struct {
	DefID               string              `json:"id"`
	Type                attribute.Type      `json:"type"`
	Placement           attribute.Placement `json:"placement,omitempty"`
	Name                string              `json:"name"`
	FullName            string              `json:"full_name,omitempty"`
	AttributeBase       string              `json:"attribute_base,omitempty"`
	CostPerPoint        fxp.Int             `json:"cost_per_point,omitempty"`
	CostAdjPercentPerSM fxp.Int             `json:"cost_adj_percent_per_sm,omitempty"`
	Thresholds          []*PoolThreshold    `json:"thresholds,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (a *AttributeDef) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	e := json.NewEncoder(&buffer)
	e.SetEscapeHTML(false)
	err := e.Encode(&a.AttributeDefData)
	return buffer.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *AttributeDef) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &a.AttributeDefData); err != nil {
		return err
	}
	for _, threshold := range a.Thresholds {
		threshold.Expression = strings.ReplaceAll(threshold.Expression, "$self", "$"+a.DefID)
	}
	return nil
}

// Clone a copy of this.
func (a *AttributeDef) Clone() *AttributeDef {
	clone := *a
	if a.Type == attribute.Pool {
		if a.Thresholds != nil {
			clone.Thresholds = make([]*PoolThreshold, len(a.Thresholds))
			for i, one := range a.Thresholds {
				clone.Thresholds[i] = one.Clone()
			}
		}
	} else {
		a.Thresholds = nil
	}
	return &clone
}

// ID returns the ID.
func (a *AttributeDef) ID() string {
	return a.DefID
}

// SetID sets the ID, sanitizing it in the process (i.e. it may be changed from what you set -- read it back if you want
// to be sure of what it gets set to.
func (a *AttributeDef) SetID(value string) {
	a.DefID = SanitizeID(value, false, ReservedIDs...)
}

// ResolveFullName returns the full name, using the short name if full name is empty.
func (a *AttributeDef) ResolveFullName() string {
	if a.FullName == "" {
		return a.Name
	}
	return a.FullName
}

// CombinedName returns the combined FullName and Name, as appropriate.
func (a *AttributeDef) CombinedName() string {
	if a.FullName == "" {
		return a.Name
	}
	if a.Name == "" || a.Name == a.FullName {
		return a.FullName
	}
	return a.FullName + " (" + a.Name + ")"
}

// IsSeparator returns true if this is actually just a separator.
func (a *AttributeDef) IsSeparator() bool {
	return a.Type == attribute.PrimarySeparator || a.Type == attribute.SecondarySeparator || a.Type == attribute.PoolSeparator
}

// Primary returns true if the base value is a non-derived value.
func (a *AttributeDef) Primary() bool {
	if a.Type == attribute.PrimarySeparator {
		return true
	}
	if a.Type == attribute.Pool || a.Placement == attribute.Secondary || a.IsSeparator() {
		return false
	}
	if a.Placement == attribute.Primary {
		return true
	}
	_, err := fxp.FromString(strings.TrimSpace(a.AttributeBase))
	return err == nil
}

// Secondary returns true if the base value is a derived value.
func (a *AttributeDef) Secondary() bool {
	if a.Type == attribute.SecondarySeparator {
		return true
	}
	if a.Type == attribute.Pool || a.Placement == attribute.Primary || a.IsSeparator() {
		return false
	}
	if a.Placement == attribute.Secondary {
		return true
	}
	_, err := fxp.FromString(strings.TrimSpace(a.AttributeBase))
	return err != nil
}

// Pool returns true if the base value is a pool value.
func (a *AttributeDef) Pool() bool {
	return a.Type == attribute.PoolSeparator || a.Type == attribute.Pool
}

// AllowsDecimal returns true if the value can have a decimal point in it.
func (a *AttributeDef) AllowsDecimal() bool {
	return a.Type == attribute.Decimal || a.Type == attribute.DecimalRef
}

// BaseValue returns the resolved base value.
func (a *AttributeDef) BaseValue(resolver eval.VariableResolver) fxp.Int {
	if a.IsSeparator() {
		return 0
	}
	return fxp.EvaluateToNumber(a.AttributeBase, resolver)
}

// ComputeCost returns the value adjusted for a cost reduction.
func (a *AttributeDef) ComputeCost(entity *Entity, value, costReduction fxp.Int, sizeModifier int) fxp.Int {
	if a.IsSeparator() {
		return 0
	}
	cost := value.Mul(a.CostPerPoint)
	if sizeModifier > 0 && a.CostAdjPercentPerSM > 0 &&
		!(a.DefID == "hp" && entity.SheetSettings.DamageProgression == progression.KnowingYourOwnStrength) {
		costReduction += fxp.From(sizeModifier).Mul(a.CostAdjPercentPerSM)
	}
	if costReduction > 0 {
		if costReduction > fxp.Eighty {
			costReduction = fxp.Eighty
		}
		cost = cost.Mul(fxp.Hundred - costReduction).Div(fxp.Hundred)
	}
	return fxp.ApplyRounding(cost, false)
}

func (a *AttributeDef) crc64(c uint64) uint64 {
	c = crc.String(c, a.DefID)
	c = crc.Byte(c, byte(a.Type))
	c = crc.Byte(c, byte(a.Placement))
	c = crc.String(c, a.Name)
	c = crc.String(c, a.FullName)
	c = crc.String(c, a.AttributeBase)
	c = crc.Number(c, a.CostPerPoint)
	c = crc.Number(c, a.CostAdjPercentPerSM)
	c = crc.Number(c, len(a.Thresholds))
	for _, one := range a.Thresholds {
		c = one.crc64(c)
	}
	return c
}
