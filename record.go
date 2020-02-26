package kintone

import (
	"encoding/json"
)

// Record ...
type Record struct {
	ID     string
	Fields Fields
}

// UnmarshalJSON ...
func (r *Record) UnmarshalJSON(data []byte) error {

	type Raw struct {
		Type  string           `json:"type"`
		Value *json.RawMessage `json:"value"`
	}
	var raws map[string]Raw

	if err := json.Unmarshal(data, &raws); err != nil {
		return err
	}

	fs := make(Fields)

	for code, raw := range raws {
		if raw.Type == FieldTypeRecordNumber {
			var id string
			err := json.Unmarshal(*raw.Value, &id)
			if err != nil {
				return err
			}
			r.ID = id
			continue
		}

		var f Field
		var err error

		switch raw.Type {
		case FieldTypeSingleLineText:
			var _f SingleLineTextField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeMultiLineText:
			var _f MultiLineTextField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeRichText:
			var _f RichTextField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeRadioButton:
			var _f RadioButtonField
			// valueの値としてnullが入ってくる可能性があるため以下のハンドリングが必要
			if raw.Value != nil {
				err = json.Unmarshal(*raw.Value, &_f)
			}
			f = _f
		case FieldTypeLink:
			var _f LinkField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeSingleSelect:
			var _f SingleSelectField
			// valueの値としてnullが入ってくる可能性があるため以下のハンドリングが必要
			if raw.Value != nil {
				err = json.Unmarshal(*raw.Value, &_f)
			}
			f = _f
		case FieldTypeStatus:
			var _f StatusField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeRecordNumber:
			var _f RecordNumberField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeID:
			var _f IDField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeCalc:
			var _f CalcField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeNumber:
			var _f NumberField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeCheckBox:
			var _f CheckBoxField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeMultiSelect:
			var _f MultiSelectField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeFile:
			var _f FileField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeDate:
			var _f DateField
			// valueの値としてnullが入ってくる可能性があるため以下のハンドリングが必要
			if raw.Value != nil {
				err = json.Unmarshal(*raw.Value, &_f)
			}
			f = _f
		case FieldTypeDateTime, FieldTypeCreatedDateTime, FieldTypeUpdatedDateTime:
			var _f DateTimeField
			// valueの値としてnullが入ってくる可能性があるため以下のハンドリングが必要
			if raw.Value != nil {
				err = json.Unmarshal(*raw.Value, &_f)
			}
			f = _f
		case FieldTypeTime:
			var _f TimeField
			// valueの値としてnullが入ってくる可能性があるため以下のハンドリングが必要
			if raw.Value != nil {
				err = json.Unmarshal(*raw.Value, &_f)
			}
			f = _f
		case FieldTypeUsers, FieldTypeAssignee:
			var _f UsersField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeCreator, FieldTypeModifier:
			var _f UserField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeOrganization:
			var _f OrganizationsField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeGroup:
			var _f GroupsField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeCategory:
			var _f CategoryField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeSubtable:
			var _f TableField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		case FieldTypeRevision:
			var _f RevisionField
			err = json.Unmarshal(*raw.Value, &_f)
			f = _f
		default:
			continue
		}

		if err != nil {
			return err
		}

		fs[code] = f
	}

	r.Fields = fs

	return nil
}

// NewRecord ...
func NewRecord(id string, fs Fields) *Record {
	return &Record{ID: id, Fields: fs}
}

func sliceRecords(rs []*Record, i int) [][]*Record {
	out := make([][]*Record, 0)
	for {
		if len(rs) > i {
			out = append(out, rs[:i])
		} else {
			out = append(out, rs)
		}

		if len(rs) > i {
			rs = rs[i:]
		} else {
			break
		}
	}
	return out
}
