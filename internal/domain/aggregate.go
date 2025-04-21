package domain

type PVZReportAggregate struct {
	PVZ        *PVZ                 `json:"pvz"`
	Receptions []ReceptionAggregate `json:"receptions"`
}

type ReceptionAggregate struct {
	Information ReceptionInfo `json:"info"`
	Products    []*Product    `json:"products"`
}
