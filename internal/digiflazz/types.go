package digiflazz

// TopupRequest is the JSON body Digiflazz sends to the seller API.
type TopupRequest struct {
	Username   string `json:"username"`
	Commands   string `json:"commands"`
	RefID      string `json:"ref_id"`
	HP         string `json:"hp"`
	PulsaCode  string `json:"pulsa_code"`
	Price      int64  `json:"price"`
	Sign       string `json:"sign"`
}

type Envelope struct {
	Data ResponseData `json:"data"`
}

type ResponseData struct {
	RefID    string `json:"ref_id"`
	Status   string `json:"status"`
	Code     string `json:"code"`
	HP       string `json:"hp"`
	Price    string `json:"price"`
	Message  string `json:"message"`
	Balance  string `json:"balance"`
	TrID     string `json:"tr_id"`
	RC       string `json:"rc"`
	SN       string `json:"sn"`
}
