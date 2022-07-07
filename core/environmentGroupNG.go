package core

type (
	// ModuleLicense represents a module license in the system
	EnvironmentGroupNG struct {
		Id                string `json:"_id,omitempty" bson:"_id,omitempty"`
		AccountIdentifier string `json:"accountId,omitempty" bson:"accountId,omitempty"`
	}
)
