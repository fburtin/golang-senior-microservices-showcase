package bcra

type DebtResponse struct {
	Status  int        `json:"status" example:"200"`
	Results DebtResult `json:"results"`
}

type DebtResult struct {
	Identification int64        `json:"identificacion" example:"20292456078"`
	Name           string       `json:"denominacion" example:"PERSONA EJEMPLO"`
	Periods        []DebtPeriod `json:"periodos"`
}

type DebtPeriod struct {
	Period   string       `json:"periodo" example:"202605"`
	Entities []DebtEntity `json:"entidades"`
}

type DebtEntity struct {
	Entity                    string  `json:"entidad" example:"BANCO DE GALICIA Y BUENOS AIRES S.A."`
	Situation                 int     `json:"situacion" example:"1"`
	SituationOneDate          string  `json:"fechaSit1,omitempty" example:"2006-09-30"`
	Amount                    float64 `json:"monto" example:"1847"`
	DaysPastDue               int     `json:"diasAtrasoPago" example:"0"`
	Refinancing               bool    `json:"refinanciaciones" example:"false"`
	MandatoryRecategorization bool    `json:"recategorizacionOblig" example:"false"`
	LegalSituation            bool    `json:"situacionJuridica" example:"false"`
	TechnicalWriteOff         bool    `json:"irrecDisposicionTecnica" example:"false"`
	UnderReview               bool    `json:"enRevision" example:"false"`
	JudicialProcess           bool    `json:"procesoJud" example:"false"`
}

type ErrorResponse struct {
	Status        int      `json:"status" example:"404"`
	ErrorMessages []string `json:"errorMessages" example:"No se encontró datos para la identificación ingresada."`
}
