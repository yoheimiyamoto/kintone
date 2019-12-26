package kintone

import (
	"encoding/json"
	"testing"
)

func TestFormFields(t *testing.T) {
	data := []byte(`
		{
			"チェックボックス": {
				"type": "CHECK_BOX",
				"code": "チェックボックス",
				"label": "チェックボックス",
				"noLabel": false,
				"required": false,
				"defaultValue": [
				  "sample1",
				  "sample3"
				],
				"options": {
				  "sample1": {
					"label": "sample1",
					"index": 0
				  },
				  "sample2": {
					"label": "sample2",
					"index": 2
				  },
				  "sample3": {
					"label": "sample3",
					"index": 1
				  }
				},
				"align": "horizontal"
			  }
		}
	`)

	var f FormFields

	err := json.Unmarshal(data, &f)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(f)
}

func TestFormLayouts(t *testing.T) {
	data := []byte(`
		{
			"revision": "2",
			"layout": [
				{
					"type": "ROW",
					"fields": [
						{
							"type": "SINGLE_LINE_TEXT",
							"code": "文字列__1行_",
							"size": {
								"width": "200"
							}
						},
						{
							"type": "MULTI_LINE_TEXT",
							"code": "文字列__複数行_",
							"size": {
								"width": "200",
								"innerHeight": "100"
							}
						},
						{
							"type": "LABEL",
							"label": "label",
							"size": {
								"width": "200"
							}
						},
						{
							"type": "SPACER",
							"elementId": "spacer",
							"size": {
								"width": "200",
								"height": "100"
							}
						},
						{
							"type": "HR",
							"size": {
								"width": "200"
							}
						}
					]
				},
				{
					"type": "SUBTABLE",
					"code": "テーブル",
					"fields": [
						{
							"type": "NUMBER",
							"code": "数値",
							"size": {
								"width": "200"
							}
						}
					]
				},
				{
					"type": "GROUP",
					"code": "グループ",
					"layout": [
						{
							"type": "ROW",
							"fields": [
								{
									"type": "NUMBER",
									"code": "数値2",
									"size": {
										"width": "200"
									}
								}
							]
						}
					]
				}
			]
		}
	`)

	var obj struct {
		Revision string        `json:"revision"`
		Layout   []*FormLayout `json:"layout"`
	}

	err := json.Unmarshal(data, &obj)
	if err != nil {
		t.Error(err)
		return
	}

	actual, err := json.Marshal(obj)
	if err != nil {
		t.Error(err)
		return
	}

	if !jsonEqual(data, actual) {
		t.Errorf("expected: %s, actual: %s", string(data), string(actual))
	}

}

func TestExtractCodes(t *testing.T) {
	layouts := FormLayouts{
		&FormLayout{
			Fields: []*FormLayoutField{&FormLayoutField{Code: "1"}},
			Layouts: []*FormLayout{
				&FormLayout{
					Fields: []*FormLayoutField{&FormLayoutField{Code: "2"}, &FormLayoutField{Code: "3"}},
					Layouts: []*FormLayout{
						&FormLayout{
							Fields: []*FormLayoutField{&FormLayoutField{Code: "4"}, &FormLayoutField{Code: "5"}},
						},
					},
				},
			},
		},
	}
	actual := layouts.Codes()
	expected := []string{"1", "2", "3", "4", "5"}
	if len(expected) != len(actual) {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}
}
