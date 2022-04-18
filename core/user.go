package core

type (
	// User represents an user in the system
	//todo: add more keys as needed
	User struct {
		Id       string   `json:"_id,omitempty" bson:"_id,omitempty"`
		Name     string   `json:"name,omitempty" bson:"name,omitempty"`
		Email    string   `json:"email,omitempty" bson:"email,omitempty"`
		Accounts []string `json:"accounts,omitempty" bson:"accounts,omitempty"`
		UtmInfo  UtmInfo  `json:"utmInfo,omitempty" bson:"utmInfo,omitempty"`
	}
	UtmInfo struct {
		UtmSource   string `json:"utmSource,omitempty" bson:"utmSource,omitempty"`
		UtmContent  string `json:"utmContent,omitempty" bson:"utmContent,omitempty"`
		UtmMedium   string `json:"utmMedium,omitempty" bson:"utmMedium,omitempty"`
		UtmTerm     string `json:"utmTerm,omitempty" bson:"utmTerm,omitempty"`
		UtmCampaign string `json:"utmCampaign,omitempty" bson:"utmCampaign,omitempty"`
	}
)

//func (a *User) GetIdAsString() string {
//	return a.Id.Hex()
//}
