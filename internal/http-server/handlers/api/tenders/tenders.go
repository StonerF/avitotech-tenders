package tenders

type Tender struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	ServiceType     string `json:"serviceType"`
	OrganizationId  string `json:"organizationId"`
	CreatorUsername string `json:"creatorUsername"`
	Version         int    `json:"version"`
	CreatedAt       string `json:"createdAt"`
}
type TenderResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ServiceType string `json:"serviceType"`
	Version     int    `json:"version"`
	CreatedAt   string `json:"createdAt"`
}

type Tenders interface {
	//Create(name string, description string, serviceType string, status string, organizationId string, creatorUsername string) (Tender, error)
	Get() []Tender
	Send() error
	Patch(tend Tender) (Tender, error)
}
