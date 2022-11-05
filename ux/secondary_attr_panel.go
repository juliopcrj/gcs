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

package ux

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/attribute"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

// SecondaryAttrPanel holds the contents of the secondary attributes block on the sheet.
type SecondaryAttrPanel struct {
	unison.Panel
	entity    *gurps.Entity
	targetMgr *TargetMgr
	prefix    string
	crc       uint64
}

// NewSecondaryAttrPanel creates a new secondary attributes panel.
func NewSecondaryAttrPanel(entity *gurps.Entity, targetMgr *TargetMgr) *SecondaryAttrPanel {
	p := &SecondaryAttrPanel{
		entity:    entity,
		targetMgr: targetMgr,
		prefix:    targetMgr.NextPrefix(),
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		VSpan:  2,
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
	})
	p.SetBorder(unison.NewCompoundBorder(&TitledBorder{Title: i18n.Text("Secondary Attributes")},
		unison.NewEmptyBorder(unison.Insets{
			Top:    1,
			Left:   2,
			Bottom: 1,
			Right:  2,
		})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}
	attrs := gurps.SheetSettingsFor(p.entity).Attributes
	p.crc = attrs.CRC64()
	p.rebuild(attrs)
	return p
}

func (p *SecondaryAttrPanel) rebuild(attrs *gurps.AttributeDefs) {
	focusRefKey := p.targetMgr.CurrentFocusRef()
	p.RemoveAllChildren()
	for _, def := range attrs.List(false) {
		if def.Secondary() {
			if def.Type == attribute.SecondarySeparator {
				p.AddChild(NewPageInternalHeader(def.Name, 3))
			} else {
				attr, ok := p.entity.Attributes.Set[def.ID()]
				if !ok {
					jot.Warnf("unable to locate attribute data for '%s'", def.ID())
					continue
				}
				p.AddChild(p.createPointsField(attr))
				p.AddChild(p.createValueField(def, attr))
				p.AddChild(NewPageLabel(def.CombinedName()))
			}
		}
	}
	if p.targetMgr != nil {
		if sheet := unison.Ancestor[*Sheet](p); sheet != nil {
			p.targetMgr.ReacquireFocus(focusRefKey, sheet.toolbar, sheet.scroll.Content())
		}
	}
}

func (p *SecondaryAttrPanel) createPointsField(attr *gurps.Attribute) *NonEditablePageField {
	field := NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := "[" + attr.PointCost().String() + "]"; text != f.Text {
			f.Text = text
			MarkForLayoutWithinDockable(f)
		}
		if def := attr.AttributeDef(); def != nil {
			f.Tooltip = unison.NewTooltipWithText(fmt.Sprintf(i18n.Text("Points spent on %s"), def.CombinedName()))
		}
	})
	field.Font = theme.PageFieldSecondaryFont
	return field
}

func (p *SecondaryAttrPanel) createValueField(def *gurps.AttributeDef, attr *gurps.Attribute) *DecimalField {
	field := NewDecimalPageField(p.targetMgr, p.prefix+attr.AttrID, def.CombinedName(),
		func() fxp.Int { return attr.Maximum() },
		func(v fxp.Int) { attr.SetMaximum(v) }, fxp.Min, fxp.Max, true)
	return field
}

// Sync the panel to the current data.
func (p *SecondaryAttrPanel) Sync() {
	attrs := gurps.SheetSettingsFor(p.entity).Attributes
	if crc := attrs.CRC64(); crc != p.crc {
		p.crc = crc
		p.rebuild(attrs)
		MarkForLayoutWithinDockable(p)
	}
}