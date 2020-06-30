package kintone

import (
	"encoding/json"
)

//+FormFields
type FormFields map[string]*FormField

type FormField struct {
	Label          string         `json:"label"`
	Code           string         `json:"code"`
	Type           string         `json:"type"`
	NoLabel        bool           `json:"noLabel"`
	Required       bool           `json:"required"`
	Unique         bool           `json:"unique"`
	MaxValue       string         `json:"maxValue"`
	MinValue       string         `json:"minValue"`
	MaxLength      string         `json:"maxLength"`
	MinLength      string         `json:"minLength"`
	DefaultValue   interface{}    `json:"defaultValue"` // string or []string
	DefaultTime    bool           `json:"defaultNowValue"`
	Options        Options        `json:"options"`
	Align          string         `json:"align"`
	HideExpression bool           `json:"hideExpression"`
	Digit          bool           `json:"digit"` // 数値の桁区切りを表示するかどうかの設定
	ThumbnailSize  string         `json:"thumbnailSize"`
	Protocol       string         `json:"protocol"`       // リンクの種類（WEB, CALL, MAIL）
	Format         string         `json:"format"`         // 計算フィールドの表示形式（NUMBER, NUMBER_DIGIT, DATETIME, DATE, TIME, HOUR_MINUTE, DAY_HOUR_MINUTE）
	DisplayScale   string         `json:"disaplayScale"`  // 小数点以下の表示桁数です。未設定の場合は空です。
	Unit           string         `json:"unit"`           // 単位の記号
	UnitPosition   string         `json:"unitPosition"`   // 位記号の表示位置（BEFORE or AFTER）
	Entities       []*Entity      `json:"entities"`       // 選択肢のユーザー
	ReferenceTable ReferenceTable `json:"referenceTalbe"` // 関連レコード一覧
	Lookup         Lookup         `json:"lookup"`
	OpenGroup      bool           `json:"openGroup"`
	Fields         FormFields     `json:"fields"`
	Enabled        bool           `json:"enabled"`
}

type Options []string

func (os *Options) UnmarshalJSON(data []byte) error {
	var raw map[string]struct {
		Label string `json:"label"`
		Index int    `json:"index,string"`
	}

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	// indexの最大値を取得
	var maxIndex int
	for _, v := range raw {
		if v.Index > maxIndex {
			maxIndex = v.Index
		}
	}

	args := make([]string, maxIndex+1)

	for _, v := range raw {
		args[v.Index] = v.Label
	}

	*os = Options(args)
	return nil
}

type Entity struct {
	Code string `json:"code"`
	Type string `json:"type"`
}

// ReferenceTable ... 関連レコード一覧
type ReferenceTable struct {
	RelatedApp    RelatedApp `json:"relatedApp"`
	Condition     Condition  `json:"condition"`
	FilterCond    string     `json:"filterCond"`    // 「さらに絞り込む条件」の設定
	DisplayFields []string   `json:"displayFields"` // 「表示するフィールド」に指定されたフィールド
	Sort          string     `json:"sort"`          // レコードのソートの設定です。クエリ形式で表されます。
	Size          int        `json:"size"`          // 一度に表示する最大レコード数
}

type Condition struct {
	Field        string `json:"field"`
	RelatedField string `json:"relatedField"`
}

type RelatedApp struct {
	App  string `json:"app"`
	Code string `json:"code"`
}

type Lookup struct {
	RelatedApp         RelatedApp    `json:"relatedApp"`
	RelatedKeyField    string        `json:"relatedKeyField"`
	FieldMappings      FieldMappings `json:"fieldMappings"`
	LookupPickerFields []string      `json:"lookupPickerFields"`
	FilterCond         string        `json:"filterCond"`
	Sort               string        `json:"sort"`
}

type FieldMappings struct {
	Field        string `json:"field"`
	RelatedField string `json:"relatedField"`
}

//-FormFields

//+FormLayout

type FormLayouts []*FormLayout

type FormLayout struct {
	Type    string             `json:"type"`
	Code    string             `json:"code,omitempty"` // テーブル or グループ レイアウトでのみ使用
	Fields  []*FormLayoutField `json:"fields,omitempty"`
	Layouts []*FormLayout      `json:"layout,omitempty"`
}

type FormLayoutField struct {
	Type      string `json:"type"`
	Code      string `json:"code,omitempty"`
	LabelName string `json:"label,omitempty"`     // ラベルフィールドの時のみ値が入る
	ElementID string `json:"elementId,omitempty"` // エレメントフィールドの時のみ値が入る
	Size      struct {
		Width       string `json:"width,omitempty"`
		Height      string `json:"height,omitempty"`
		InnerHeight string `json:"innerHeight,omitempty"`
	} `json:"size"`
}

// 再帰的処理

func (f FormLayouts) Codes() []string {
	var stack []string

	var extract func(layouts []*FormLayout)
	extract = func(layouts []*FormLayout) {
		for _, l := range layouts {
			for _, f := range l.Fields {
				stack = append(stack, f.Code)
			}
			extract(l.Layouts)
		}
	}

	extract(f)

	return stack
}

//-FormLayout
