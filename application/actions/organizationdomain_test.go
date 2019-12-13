package actions

import (
	"fmt"

	"github.com/silinternational/wecarry-api/models"
)

type OrganizationDomainFixtures struct {
	models.Organization
	models.Users
}

type OrganizationDomainResponse struct {
	OrganizationDomain []struct {
		OrganizationID string `json:"organizationID"`
		Domain         string `json:"domain"`
	} `json:"domain"`
}

func (as *ActionSuite) Test_CreateOrganizationDomain() {
	f := fixturesForOrganizationDomain(as)

	testDomain := "example.com"
	allFieldsQuery := `organizationID domain`
	allFieldsInput := fmt.Sprintf(`organizationID:"%s" domain:"%s"`,
		f.Organizations[0].UUID.String(), testDomain)

	query := fmt.Sprintf("mutation{domain: createOrganizationDomain(input: {%s}) {%s}}",
		allFieldsInput, allFieldsQuery)
	var resp OrganizationDomainResponse
	err := as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)

	as.Equal(1, len(resp.OrganizationDomain), "wrong number of domains in response")
	as.Equal(testDomain, resp.OrganizationDomain[0].Domain, "received wrong domain")
	as.Equal(f.Organizations[0].UUID.String(), resp.OrganizationDomain[0].OrganizationID, "received wrong org ID")

	var orgs models.Organizations
	err = as.DB.Eager().Where("name = ?", f.Organizations[0].Name).All(&orgs)
	as.NoError(err)

	as.GreaterOrEqual(1, len(orgs), "no Organization found")
	as.Equal(1, len(orgs[0].OrganizationDomains), "wrong number of domains in DB")
	as.Equal(testDomain, orgs[0].OrganizationDomains[0].Domain, "wrong domain in DB")
}