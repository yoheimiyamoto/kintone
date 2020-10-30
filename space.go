package kintone

import "encoding/json"

type Space struct {
	ID           int    `json:"id,string"`
	Name         string `json:"name"`
	IsPrivate    bool   `json:"isPrivate"`
	MemberCount  int    `json:"memberCount,string"`
	Body         string `json:"body"`
	AttachedApps []*App `json:"attachedApps"`
}

type App struct {
	AppID       int            `json:"appID,string"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	CreatedAt   *DateTimeField `json:"createdAt"`
	Creator     *Entity        `json:"creator"`
	Modifier    *Entity        `json:"modifier"`
	UpdatedAt   *DateTimeField `json:"modifiedAt"`
}

type CreateSpace struct {
	TemplateID int                  `json:"id"`
	Name       string               `json:"name"`
	Members    []*CreateSpaceMember `json:"members"`
	IsPrivate  bool                 `json:"isPrivate"`
}

type CreateSpaceMember struct {
	EntityType string
	Code       string
	IsAdmin    bool
}

func (r *CreateSpaceMember) MarshalJSON() ([]byte, error) {
	type Entity struct {
		EntityType string `json:"type"`
		Code       string `json:"code"`
	}

	raw := struct {
		Entity  *Entity `json:"entity"`
		IsAdmin bool    `json:"isAdmin"`
	}{
		Entity:  &Entity{r.EntityType, r.Code},
		IsAdmin: true,
	}

	return json.Marshal(&raw)
}
