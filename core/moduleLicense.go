package core

type (
	// ModuleLicense represents a module license in the system
	ModuleLicense struct {
		Id                string `json:"_id,omitempty" bson:"_id,omitempty"`
		AccountIdentifier string `json:"accountIdentifier,omitempty" bson:"accountIdentifier,omitempty"`
		ModuleType        string `json:"moduleType,omitempty" bson:"moduleType,omitempty"`
		Edition           string `json:"edition,omitempty" bson:"edition,omitempty"`
		LicenseType       string `json:"licenseType,omitempty" bson:"licenseType,omitempty"`
		Status            string `json:"status,omitempty" bson:"status,omitempty"`
		StartTime         int64  `json:"startTime,omitempty" bson:"startTime,omitempty"`
		ExpiryTime        int64  `json:"expiryTime,omitempty" bson:"expiryTime,omitempty"`
		LastUpdatedAt     int64  `json:"lastUpdatedAt,omitempty" bson:"lastUpdatedAt,omitempty"`
		CreatedAt         int64  `json:"createdAt,omitempty" bson:"createdAt,omitempty"`

		//CD
		CDLicenseType    string `json:"cdLicenseType,omitempty" bson:"cdLicenseType,omitempty"`
		ServiceInstances int    `json:"serviceInstances,omitempty" bson:"serviceInstances,omitempty"`
		Workloads        int    `json:"workloads,omitempty" bson:"workloads,omitempty"`

		//CI
		NumberOfCommitters int `json:"numberOfCommitters,omitempty" bson:"numberOfCommitters,omitempty"`

		//FF-CF
		NumberOfUsers      int   `json:"numberOfUsers,omitempty" bson:"numberOfUsers,omitempty"`
		NumberOfClientMAUs int64 `json:"numberOfClientMAUs,omitempty" bson:"numberOfClientMAUs,omitempty"`

		//CCM-CE
		SpendLimit int64 `json:"spendLimit,omitempty" bson:"spendLimit,omitempty"`

		//SRM
		NumberOfServices int `json:"numberOfServices,omitempty" bson:"numberOfServices,omitempty"`

		//CHAOS
		TotalChaosExperimentRuns int `json:"totalChaosExperimentRuns,omitempty" bson:"totalChaosExperimentRuns,omitempty"`

		//STO
		NumberOfDevelopers int `json:"numberOfDevelopers,omitempty" bson:"numberOfDevelopers,omitempty"`
	}
)
