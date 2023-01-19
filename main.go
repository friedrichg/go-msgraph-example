package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	abstractions "github.com/microsoft/kiota-abstractions-go"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	az "github.com/microsoft/kiota-authentication-azure-go"
)

var (
	validhosts = []string{"graph.microsoft.com", "graph.microsoft.us", "dod-graph.microsoft.us", "graph.microsoft.de", "microsoftgraph.chinacloudapi.cn", "canary.graph.microsoft.com"}
	scopes     = []string{"https://graph.microsoft.com/.default"}
)

const (
	listApplicationURI = "https://graph.microsoft.com/v1.0/applications"
)

type application struct {
	display      string
	client       *http.Client
	authProvider *az.AzureIdentityAuthenticationProvider
}

type azureTransport struct {
	*http.Transport
	app *application
}

func (t *azureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ri := abstractions.NewRequestInformation()
	ri.SetUri(*req.URL)
	err := t.app.authProvider.AuthenticateRequest(req.Context(), ri, nil)
	if err != nil {
		return nil, err
	}
	for _, k := range ri.Headers.ListKeys() {
		req.Header.Add(k, strings.Join(ri.Headers.Get(k), ","))
	}

	return t.Transport.RoundTrip(req)
}

func main() {

	a := application{
		display: os.Getenv("DISPLAY"),
	}
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	a.authProvider, err = az.NewAzureIdentityAuthenticationProviderWithScopesAndValidHosts(cred, scopes, validhosts)
	if err != nil {
		log.Fatalf("failed to obtain an authProvider: %v", err)
	}

	a.client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &azureTransport{
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
			app:       &a,
		},
	}

	err = a.getApplication()
	if err != nil {
		log.Fatalf("failed to getApplication: %v", err)
	}
}

type applicationListResponse struct {
	Value []applicationDescription
}

type applicationDescription struct {
	ID                     string                       `json:"id,omitempty"`
	AppID                  string                       `json:"appId,omitempty"`
	DisplayName            string                       `json:"displayName,omitempty"`
	GroupMembershipClaims  string                       `json:"groupMembershipClaims,omitempty"`
	SignInAudience         string                       `json:"signInAudience,omitempty"`
	Web                    *web                         `json:"web,omitempty"`
	RequiredResourceAccess []requiredResourceAccessItem `json:"requiredResourceAccess,omitempty"`
}

type web struct {
	RedirectUris []string `json:"redirectUris,omitempty"`
}

type requiredResourceAccessItem struct {
	ResourceAppId  string               `json:"resourceAppId,omitempty"`
	ResourceAccess []resourceAccessItem `json:"resourceAccess,omitempty"`
}

type resourceAccessItem struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

func (a *application) getApplication() error {
	base, _ := url.Parse(listApplicationURI)
	params := url.Values{}
	params.Add("$filter", fmt.Sprintf("displayName eq '%v'", a.display))
	base.RawQuery = params.Encode()
	req, err := http.NewRequest(http.MethodGet, base.String(), nil)
	if err != nil {
		return err
	}
	response, err := a.client.Do(req)
	if err != nil {
		return err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("status error: %v, %s", response.StatusCode, responseData)
	}
	var responseObject applicationListResponse
	err = json.Unmarshal(responseData, &responseObject)
	if err != nil {
		return err
	}
	var found bool
	for _, app := range responseObject.Value {
		log.Printf("app exists, app id: %v\n", app.AppID)
		found = true
	}
	if !found {
		log.Printf("app not found\n")
	}
	return nil

}
