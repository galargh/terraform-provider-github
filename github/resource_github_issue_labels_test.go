package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccGithubIssueLabels(t *testing.T) {
	t.Run("authoritatively overtakes existing labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		empty := []map[string]interface{}{}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					// 0. Check if labels start from a clean slate
					{
						Config: testAccGithubIssueLabelsConfig(randomID, false, empty),
						Check:  resource.TestCheckResourceAttr("github_issue_labels.test", "label.#", "0"),
					},
					// 1. Check if a label can be created
					{
						Config: testAccGithubIssueLabelsConfig(randomID, false, append(empty, map[string]interface{}{
							"name":        "foo",
							"color":       "000000",
							"description": "foo",
						})),
						Check: resource.TestCheckResourceAttr("github_issue_labels.test", "label.#", "1"),
					},
					// 2. Check if a label can be recreated
					{
						Config: testAccGithubIssueLabelsConfig(randomID, false, append(empty, map[string]interface{}{
							"name":        "Foo",
							"color":       "000000",
							"description": "foo",
						})),
						Check: resource.TestCheckResourceAttr("github_issue_labels.test", "label.#", "1"),
					},
					// 3. Check if multiple labels can be created
					{
						Config: testAccGithubIssueLabelsConfig(randomID, false, append(empty,
							map[string]interface{}{
								"name":        "Foo",
								"color":       "000000",
								"description": "foo",
							},
							map[string]interface{}{
								"name":        "bar",
								"color":       "000000",
								"description": "bar",
							}, map[string]interface{}{
								"name":        "baz",
								"color":       "000000",
								"description": "baz",
							})),
						Check: resource.TestCheckResourceAttr("github_issue_labels.test", "label.#", "3"),
					},
					// 4. Check if labels can be destroyed
					{
						Config: testAccGithubIssueLabelsConfig(randomID, false, empty),
						Check:  resource.TestCheckResourceAttr("github_issue_labels.test", "label.#", "0"),
					},
					// 5. Check if the default labels were left untouched (note the switch to authoritative mode)
					{
						Config:             testAccGithubIssueLabelsConfig(randomID, true, empty),
						ExpectNonEmptyPlan: true,
					},
					// 6. Check if the default labels can be destroyed
					{
						Config: testAccGithubIssueLabelsConfig(randomID, true, empty),
						Check:  resource.TestCheckResourceAttr("github_issue_labels.test", "label.#", "0"),
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})
}

func testAccGithubIssueLabelsConfig(randomId string, authoritative bool, labels []map[string]interface{}) string {
	dynamic := ""
	for _, label := range labels {
		dynamic += fmt.Sprintf(`
			label {
				name = "%s"
				color = "%s"
				description = "%s"
			}
		`, label["name"], label["color"], label["description"])
	}

	return fmt.Sprintf(`
		resource "github_repository" "test" {
			name = "tf-acc-test-%s"
			auto_init = true
		}

		resource "github_issue_labels" "test" {
			repository = github_repository.test.id

			authoritative = %v

			%s
		}
	`, randomId, authoritative, dynamic)
}
