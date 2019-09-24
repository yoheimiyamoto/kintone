package kintone

import (
	"encoding/json"
	"strings"
	"time"
)

// FieldType ...
const (
	FieldTypeSingleLineText  = "SINGLE_LINE_TEXT"
	FieldTypeMultiLineText   = "MULTI_LINE_TEXT"
	FieldTypeRichText        = "RICH_TEXT"
	FieldTypeNumber          = "NUMBER"
	FieldTypeCalc            = "CALC"
	FieldTypeCheckBox        = "CHECK_BOX"
	FieldTypeRadioButton     = "RADIO_BUTTON"
	FieldTypeSingleSelect    = "DROP_DOWN"
	FieldTypeMultiSelect     = "MULTI_SELECT"
	FieldTypeFile            = "FILE"
	FieldTypeLink            = "LINK"
	FieldTypeDate            = "DATE"
	FieldTypeTime            = "TIME"
	FieldTypeDateTime        = "DATETIME"
	FieldTypeUsers           = "USER_SELECT"
	FieldTypeOrganization    = "ORGANIZATION_SELECT"
	FieldTypeGroup           = "GROUP_SELECT"
	FieldTypeCategory        = "CATEGORY"
	FieldTypeStatus          = "STATUS"
	FieldTypeAssignee        = "STATUS_ASSIGNEE"
	FieldTypeRecordNumber    = "RECORD_NUMBER"
	FieldTypeCreator         = "CREATOR"
	FieldTypeCreatedDateTime = "CREATED_TIME"
	FieldTypeModifier        = "MODIFIER"
	FieldTypeUpdatedDateTime = "UPDATED_TIME"
	FieldTypeSubtable        = "SUBTABLE"
	FieldTypeID              = "__ID__"
	FieldTypeRevision        = "__REVISION__"
)

// Field ...
type Field interface {
	// String() string
}

// Fields ...
type Fields map[string]Field

// MarshalJSON ...
func (f Fields) MarshalJSON() ([]byte, error) {
	obj := make(map[string]struct {
		Value Field `json:"value"`
	})

	for k, v := range f {
		obj[k] = struct {
			Value Field `json:"value"`
		}{v}
	}

	return json.Marshal(obj)
}

//+string

// SingleLineTextField ...
type SingleLineTextField string

// MultiLineTextField ...
type MultiLineTextField string

// RichTextField ...
type RichTextField string

// RadioButtonField ...
type RadioButtonField string

// SingleSelectField ...
type SingleSelectField string

// LinkField ...
type LinkField string

// StatusField ...
type StatusField string

// RecordNumberField ...
type RecordNumberField string

// IDField ...
type IDField string

// CalcField ...
type CalcField string

// RevisionField ...
type RevisionField string

// NumberField ...
type NumberField string

//-String

//+TextsField

// TextsField ...
type TextsField []string

func (f TextsField) String() string {
	return strings.Join([]string(f), ",")
}

// CheckBoxField ...
type CheckBoxField TextsField

// MultiSelectField ...
type MultiSelectField TextsField

// CategoryField ...
type CategoryField TextsField

//-TextsField

//+DateField

// DateField ...
type DateField time.Time

func (f DateField) String() string {
	return time.Time(f).Format("2006-01-02")
}

// UnmarshalJSON ...
func (f *DateField) UnmarshalJSON(data []byte) error {
	var v string
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	t, _ := time.Parse("2006-01-02", v)

	*f = DateField(t)
	return nil
}

// MarshalJSON ...
func (f DateField) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

// NewDateField ...
func NewDateField(year, month, day int) *DateField {
	t := DateField(time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC))
	return &t
}

//-DateField

//+DateTimeField

// DateTimeField ...
type DateTimeField time.Time

func (f DateTimeField) String() string {
	return time.Time(f).Format(time.RFC3339)
}

// UnmarshalJSON ...
func (f *DateTimeField) UnmarshalJSON(data []byte) error {
	var v string
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	t, _ := time.Parse(time.RFC3339, v)

	*f = DateTimeField(t)
	return nil
}

// MarshalJSON ...
func (f DateTimeField) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func newDateTimeField(year, month, day, hour, min int) *DateTimeField {
	t := DateTimeField(time.Date(year, time.Month(month), day, hour, min, 0, 0, time.UTC))
	return &t
}

//-DateTimeField

//+TimeField

// TimeField ...
type TimeField string

func newTimeField(hour, min int) TimeField {
	t := time.Date(1, time.January, 1, hour, min, 0, 0, time.UTC)
	return TimeField(t.Format("15:04"))
}

//-TimeField

//+FileField

// FileField ...
type FileField []*File

func newFileField(value interface{}) (FileField, error) {
	data, _ := json.Marshal(value)
	var field FileField
	err := json.Unmarshal(data, &field)
	if err != nil {
		return nil, err
	}
	return field, nil
}

func (f FileField) String() string {
	var args []string
	for _, _f := range f {
		args = append(args, _f.String())
	}
	return strings.Join(args, ",")
}

// File ...
type File struct {
	ContentType string `json:"contentType"`
	FileKey     string `json:"fileKey"`
	Name        string `json:"name"`
	Size        uint64 `json:"size,string"`
}

func (f File) String() string {
	return string(f.Name)
}

//-FileField

//+CodeField

// CodeField ...
type CodeField struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (c CodeField) String() string {
	return c.Name
}

// UserField ...
type UserField CodeField

// CodeFields ...
type CodeFields []*CodeField

func (f CodeFields) String() string {
	args := make([]string, len(f))
	for i, v := range f {
		args[i] = v.Name
	}
	return strings.Join(args, ",")
}

// UsersField ...
type UsersField CodeFields

// OrganizationsField ...
type OrganizationsField CodeFields

// GroupsField ...
type GroupsField CodeFields

//-CodeField

//+Table

// TableField ...
type TableField []*Record

func (f *TableField) String() string {
	return ""
}

// MarshalJSON ...
func (f TableField) MarshalJSON() ([]byte, error) {
	type Raw struct {
		ID    string `json:"id"`
		Value Fields `json:"value"`
	}
	raws := make([]*Raw, len(f))
	for i, r := range f {
		raws[i] = &Raw{r.ID, r.Fields}
	}
	return json.Marshal(&raws)
}

// UnmarshalJSON ...
func (f *TableField) UnmarshalJSON(data []byte) error {
	type Raw struct {
		ID    string      `json:"id"`
		Value interface{} `json:"value"`
	}
	var raws []*Raw
	err := json.Unmarshal(data, &raws)
	if err != nil {
		return err
	}

	rs := make([]*Record, len(raws))
	for i, raw := range raws {
		data, err = json.Marshal(raw.Value)
		if err != nil {
			return err
		}

		var r Record
		err = json.Unmarshal(data, &r)
		if err != nil {
			return err
		}
		r.ID = raw.ID
		rs[i] = &r
	}
	*f = TableField(rs)
	return nil
}

//-Table
