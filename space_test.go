package kintone

import (
	"encoding/json"
	"testing"
)

func TestSpace(t *testing.T) {
	data := []byte(`
	{
		"id": "1",
		"name": "全体連絡スペース",
		"defaultThread": "3",
		"isPrivate": true,
		"creator": {
			"code": "tanaka",
			"name": "田中太郎"
		},
		"modifier": {
			"code": "tanaka",
			"name": "田中太郎"
		},
		"memberCount": 10,
		"coverType": "PRESET",
		"coverKey": "GREEN",
		"coverUrl": "https://*******/green.jpg",
		"body": "<b>全体</b>のスペースです。",
		"useMultiThread": true,
		"isGuest": false,
		"attachedApps": [
			{
				"threadId": "1",
				"appId": "1",
				"code": "TASK",
				"name": "タスク管理",
				"description": "タスクを管理するアプリです。",
				"createdAt": "2012-02-03T09:22:00Z",
				"creator": {
					"name": "佐藤昇",
					"code": "sato"
				},
				"modifiedAt": "2012-04-15T10:08:00Z",
				"modifier": {
					"name": "佐藤昇",
					"code": "sato"
				}
			},
			{
				"threadId": "3",
				"appId": "10",
				"code": "",
				"name": "アンケートフォーム",
				"description": "アンケートアプリです。",
				"createdAt": "2012-02-03T09:22:00Z",
				"creator": {
					 "name": "佐藤昇",
					 "code": "sato"
				},
				"modifiedAt": "2012-04-15T10:08:00Z",
					 "modifier": {
					 "name": "佐藤昇",
					 "code": "sato"
				}
			},
			{
				"threadId": "3",
				"appId": "11",
				"code": "",
				"name": "日報",
				"description": "日報アプリです。",
				"createdAt": "2012-02-03T09:22:00Z",
				"creator": {
					 "name": "加藤美咲",
					 "code": "kato"
				},
				"modifiedAt": "2012-04-15T10:08:00Z",
				"modifier": {
					 "name": "加藤美咲",
					 "code": "kato"
				}
			}
		]
	}
`)

	var f Space

	err := json.Unmarshal(data, &f)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(f)
}

func TestUnmarshalCreateSpace(t *testing.T) {
	s := CreateSpace{
		TemplateID: 1,
		Name:       "test",
		Members: []*CreateSpaceMember{
			&CreateSpaceMember{
				EntityType: "USER",
				Code:       "mycode",
				IsAdmin:    true,
			},
		},
	}
	data, err := json.Marshal(&s)
	if err != nil {
		t.Error(err)
		return
	}

	actual := string(data)
	expected := `{"id":1,"name":"test","members":[{"entity":{"type":"mycode","value":"mycode"},"isAdmin":true}]}`

	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
		return
	}
}
