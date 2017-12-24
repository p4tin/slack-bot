package app

type ResponseRtmStart struct {
	Ok    bool         `json:"ok"`
	Error string       `json:"error"`
	Url   string       `json:"url"`
	Self  ResponseSelf `json:"self"`
}

type ResponseSelf struct {
	Id string `json:"id"`
}

type Message struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
	User    string `json: "user"`
}

type UserInfo struct {
	Ok   bool `json:"ok"`
	User struct {
		Color             string `json:"color"`
		Deleted           bool   `json:"deleted"`
		ID                string `json:"id"`
		IsAdmin           bool   `json:"is_admin"`
		IsBot             bool   `json:"is_bot"`
		IsOwner           bool   `json:"is_owner"`
		IsPrimaryOwner    bool   `json:"is_primary_owner"`
		IsRestricted      bool   `json:"is_restricted"`
		IsUltraRestricted bool   `json:"is_ultra_restricted"`
		Name              string `json:"name"`
		Profile           struct {
			Email              string `json:"email"`
			FirstName          string `json:"first_name"`
			Image192           string `json:"image_192"`
			Image24            string `json:"image_24"`
			Image32            string `json:"image_32"`
			Image48            string `json:"image_48"`
			Image72            string `json:"image_72"`
			ImageOriginal      string `json:"image_original"`
			LastName           string `json:"last_name"`
			Phone              string `json:"phone"`
			RealName           string `json:"real_name"`
			RealNameNormalized string `json:"real_name_normalized"`
			Skype              string `json:"skype"`
			Title              string `json:"title"`
		} `json:"profile"`
		RealName string      `json:"real_name"`
		Status   interface{} `json:"status"`
		TeamID   string      `json:"team_id"`
		Tz       string      `json:"tz"`
		TzLabel  string      `json:"tz_label"`
		TzOffset int         `json:"tz_offset"`
	} `json:"user"`
}

type Queue struct {
	Name      string
	Paused    bool
	MaxDepth  int
	LastDepth int
}
