package MultiplayerServer

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type qosServer struct {
	ServerUrl string
	Region    string
}

type listPartyQosServersResponse struct {
	QosServers []qosServer
	PageSize   int
}

func ListPartyQosServers(w http.ResponseWriter, _ *http.Request) {
	/*shared.RespondOK(
		&w,
		listPartyQosServersResponse{
			QosServers: []qosServer{
				{ServerUrl: "mpsqosprod.westus.cloudapp.azure.com", Region: "WestUs"},
				{ServerUrl: "mpsqosprod.centralus.cloudapp.azure.com", Region: "CentralUs"},
				{ServerUrl: "mpsqosprod.westeurope.cloudapp.azure.com", Region: "WestEurope"},
				{ServerUrl: "mpsqosprod.southcentralus.cloudapp.azure.com", Region: "SouthCentralUs"},
				{ServerUrl: "mpsqosprod.northeurope.cloudapp.azure.com", Region: "NorthEurope"},
				{ServerUrl: "mpsqosprod.northcentralus.cloudapp.azure.com", Region: "NorthCentralUs"},
				{ServerUrl: "mpsqosprod.eastus2.cloudapp.azure.com", Region: "EastUs2"},
				{ServerUrl: "mpsqosprod.eastus.cloudapp.azure.com", Region: "EastUs"},
				{ServerUrl: "mpsqosprod.brazilsouth.cloudapp.azure.com", Region: "BrazilSouth"},
				{ServerUrl: "mpsqosprod.australiaeast.cloudapp.azure.com", Region: "AustraliaEast"},
				{ServerUrl: "mpsqosprod.australiasoutheast.cloudapp.azure.com", Region: "AustraliaSoutheast"},
				{ServerUrl: "mpsqosprod.eastasia.cloudapp.azure.com", Region: "EastAsia"},
				{ServerUrl: "mpsqosprod.japanwest.cloudapp.azure.com", Region: "JapanWest"},
				{ServerUrl: "mpsqosprod.japaneast.cloudapp.azure.com", Region: "JapanEast"},
				{ServerUrl: "mpsqosprod.southeastasia.cloudapp.azure.com", Region: "SoutheastAsia"},
				{ServerUrl: "mpsqosprod.southafricanorth.cloudapp.azure.com", Region: "SouthAfricaNorth"},
				{ServerUrl: "mpsqosprod.uaenorth.cloudapp.azure.com", Region: "UaeNorth"},
			},
			PageSize: 17,
		},
	)*/
	shared.RespondNotAvailable(&w)
}
