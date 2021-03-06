package acceptance_test

import (
	"crypto/tls"
	"fmt"
	"lib/models"
	"lib/policy_client"
	"math/rand"
	"net/http"
	"strings"

	"code.cloudfoundry.org/lager/lagertest"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/warrant"
)

var _ = Describe("space developer policy configuration", func() {
	var (
		appA       string
		appB       string
		orgName    string
		spaceNameA string
		spaceNameB string
		prefix     string
		w          warrant.Warrant

		policyClient *policy_client.ExternalClient
	)

	BeforeEach(func() {
		policyClient = policy_client.NewExternal(lagertest.NewTestLogger("test"),
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			},
			fmt.Sprintf("https://%s", config.ApiEndpoint),
		)

		prefix = testConfig.Prefix
		appA = fmt.Sprintf("appA-%d", rand.Int31())
		appB = fmt.Sprintf("appB-%d", rand.Int31())

		AuthAsAdmin()

		orgName = prefix + "org"
		Expect(cf.Cf("create-org", orgName).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", orgName).Wait(Timeout_Push)).To(gexec.Exit(0))

		spaceNameA = prefix + "space-A"
		Expect(cf.Cf("create-space", spaceNameA).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", orgName, "-s", spaceNameA).Wait(Timeout_Push)).To(gexec.Exit(0))

		pushProxy(appA)

		spaceNameB = prefix + "space-B"
		Expect(cf.Cf("create-space", spaceNameB).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", orgName, "-s", spaceNameB).Wait(Timeout_Push)).To(gexec.Exit(0))

		pushProxy(appB)
		w = warrant.New(warrant.Config{
			Host:          fmt.Sprintf("https://uaa.%s", config.AppsDomain),
			SkipVerifySSL: true,
		})

		uaaAdminClientToken, err := w.Clients.GetToken("admin", testConfig.AdminSecret)
		Expect(err).NotTo(HaveOccurred())

		user := ensureUserExists(w, "space-developer", "password", uaaAdminClientToken)
		group := ensureGroupExists(w, "network.write", uaaAdminClientToken)

		err = w.Groups.AddMember(group.ID, user.ID, uaaAdminClientToken)
		Expect(err).To(Or(BeNil(), BeAssignableToTypeOf(warrant.DuplicateResourceError{})))
	})

	AfterEach(func() {
		By("logging in as admin and deleting the org", func() {
			Expect(cf.Cf("auth", config.AdminUser, config.AdminPassword).Wait(Timeout_Push)).To(gexec.Exit(0))
			Expect(cf.Cf("delete-org", orgName, "-f").Wait(Timeout_Push)).To(gexec.Exit(0))
		})
	})

	Describe("space developer with network.write scope", func() {
		It("can create, list, and delete network policies in spaces they have access to", func(done Done) {
			By("setting space roles", func() {
				Expect(cf.Cf("set-space-role", "space-developer", orgName, spaceNameA, "SpaceDeveloper").Wait(Timeout_Push)).To(gexec.Exit(0))
				Expect(cf.Cf("set-space-role", "space-developer", orgName, spaceNameB, "SpaceDeveloper").Wait(Timeout_Push)).To(gexec.Exit(0))
			})

			var spaceDevUserToken string
			By("logging in and getting the space developer user token", func() {
				Expect(cf.Cf("auth", "space-developer", "password").Wait(Timeout_Push)).To(gexec.Exit(0))
				session := cf.Cf("oauth-token")
				Expect(session.Wait(Timeout_Push)).To(gexec.Exit(0))
				spaceDevUserToken = strings.TrimSpace(string(session.Out.Contents()))
			})

			var appAGUID, appBGUID string
			By("getting the app guids", func() {
				Expect(cf.Cf("target", "-o", orgName, "-s", spaceNameA).Wait(Timeout_Push)).To(gexec.Exit(0))
				session := cf.Cf("app", appA, "--guid")
				Expect(session.Wait(Timeout_Push)).To(gexec.Exit(0))
				appAGUID = strings.TrimSpace(string(session.Out.Contents()))

				Expect(cf.Cf("target", "-o", orgName, "-s", spaceNameB).Wait(Timeout_Push)).To(gexec.Exit(0))
				session = cf.Cf("app", appB, "--guid")
				Expect(session.Wait(Timeout_Push)).To(gexec.Exit(0))
				appBGUID = strings.TrimSpace(string(session.Out.Contents()))
			})

			By("creating a policy", func() {
				err := policyClient.AddPolicies(spaceDevUserToken, []models.Policy{
					models.Policy{
						Source: models.Source{
							ID: appAGUID,
						},
						Destination: models.Destination{
							ID:       appBGUID,
							Port:     1234,
							Protocol: "tcp",
						},
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			By("listing policies", func() {
				expectedPolicies := []models.Policy{
					models.Policy{
						Source: models.Source{
							ID: appAGUID,
						},
						Destination: models.Destination{
							ID:       appBGUID,
							Port:     1234,
							Protocol: "tcp",
						},
					},
				}
				policies, err := policyClient.GetPolicies(spaceDevUserToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(policies).To(Equal(expectedPolicies))
			})

			By("deleting the policy", func() {
				err := policyClient.DeletePolicies(spaceDevUserToken, []models.Policy{
					models.Policy{
						Source: models.Source{
							ID: appAGUID,
						},
						Destination: models.Destination{
							ID:       appBGUID,
							Port:     1234,
							Protocol: "tcp",
						},
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			close(done)
		}, 60 /* <-- overall spec timeout in seconds */)
	})
})

func ensureGroupExists(client warrant.Warrant, name, token string) warrant.Group {
	_, err := client.Groups.Create(name, token)
	Expect(err).To(Or(BeNil(), BeAssignableToTypeOf(warrant.DuplicateResourceError{})))

	groups, err := client.Groups.List(warrant.Query{Filter: fmt.Sprintf(`displayName eq %q`, name)}, token)
	Expect(err).NotTo(HaveOccurred())

	return groups[0]
}

func ensureUserExists(warrantClient warrant.Warrant, username, password, token string) warrant.User {
	createUserSession := cf.Cf("create-user", username, password)
	Expect(createUserSession.Wait(Timeout_Push)).To(gexec.Exit(0))

	users, err := warrantClient.Users.List(warrant.Query{Filter: fmt.Sprintf(`userName eq %q`, username)}, token)
	Expect(err).NotTo(HaveOccurred())

	return users[0]
}
