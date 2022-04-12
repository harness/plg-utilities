package core

type (
	// Account represents an account in the system
	//todo: add more keys as needed
	Account struct {
		Id             string      `json:"_id,omitempty" bson:"_id,omitempty"`
		CompanyName    string      `json:"companyName,omitempty" bson:"companyName,omitempty"`
		NextGenEnabled bool        `json:"nextGenEnabled,omitempty" bson:"nextGenEnabled,omitempty"`
		AccountName    string      `json:"accountName,omitempty" bson:"accountName,omitempty"`
		AccountKey     string      `json:"accountKey,omitempty" bson:"accountKey,omitempty"`
		LicenseInfo    LicenseInfo `json:"licenseInfo,omitempty" bson:"licenseInfo,omitempty"`
		//EncryptedLicenseInfo primitive.Binary `json:"encryptedLicenseInfo,omitempty" bson:"encryptedLicenseInfo,omitempty"`
		DefaultExperience string `json:"defaultExperience,omitempty" bson:"defaultExperience,omitempty"`
	}

	LicenseInfo struct {
		AccountType   string `json:"accountType,omitempty" bson:"accountType,omitempty"`
		AccountStatus string `json:"accountStatus,omitempty" bson:"accountStatus,omitempty"`
		ExpiryTime    int64  `json:"expiryTime,omitempty" bson:"expiryTime,omitempty"`
		LicenseUnits  int    `json:"licenseUnits,omitempty" bson:"licenseUnits,omitempty"`
	}
)

//func (a *Account) GetAccountIdAsString() string {
//	return a.Id.Hex()
//}
