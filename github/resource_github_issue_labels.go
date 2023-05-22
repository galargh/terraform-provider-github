package github

import (
	"context"
	"log"
	"strings"

	"github.com/google/go-github/v52/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubIssueLabels() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubIssueLabelsCreateOrDeleteOrUpdate,
		Read:   resourceGithubIssueLabelsRead,
		Update: resourceGithubIssueLabelsCreateOrDeleteOrUpdate,
		Delete: resourceGithubIssueLabelsCreateOrDeleteOrUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub repository.",
			},
			"label": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of labels",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the label.",
						},
						"color": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A 6 character hex code, without the leading '#', identifying the color of the label.",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A short description of the label.",
						},
						"url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The URL to the issue label.",
						},
					},
				},
			},
		},
	}
}

func resourceGithubIssueLabelsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repository := d.Id()

	log.Printf("[DEBUG] Reading GitHub issue labels for %s/%s", owner, repository)

	labels, err := githubIssueLabels(client, owner, repository)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Found %d GitHub issue labels for %s/%s", len(labels), owner, repository)
	log.Printf("[DEBUG] Labels: %v", labels)

	err = d.Set("repository", repository)
	if err != nil {
		return err
	}

	err = d.Set("label", labels)
	if err != nil {
		return err
	}

	return nil
}

func resourceGithubIssueLabelsCreateOrDeleteOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repository := d.Get("repository").(string)
	ctx := context.WithValue(context.Background(), ctxId, repository)

	o, n := d.GetChange("label")

	log.Printf("[DEBUG] Updating GitHub issue labels for %s/%s", owner, repository)
	log.Printf("[DEBUG] Old labels: %v", o)
	log.Printf("[DEBUG] New labels: %v", n)

	oMap := make(map[string]map[string]interface{})
	nMap := make(map[string]map[string]interface{})
	for _, raw := range o.(*schema.Set).List() {
		m := raw.(map[string]interface{})
		name := strings.ToLower(m["name"].(string))
		oMap[name] = m
	}
	for _, raw := range n.(*schema.Set).List() {
		m := raw.(map[string]interface{})
		name := strings.ToLower(m["name"].(string))
		nMap[name] = m
	}

	// create
	for name, n := range nMap {
		if _, ok := oMap[name]; !ok {
			log.Printf("[DEBUG] Creating GitHub issue label %s/%s/%s", owner, repository, name)

			_, _, err := client.Issues.CreateLabel(ctx, owner, repository, &github.Label{
				Name:        github.String(n["name"].(string)),
				Color:       github.String(n["color"].(string)),
				Description: github.String(n["description"].(string)),
			})
			if err != nil {
				return err
			}
		}
	}

	// delete
	for name, o := range oMap {
		if _, ok := nMap[name]; !ok {
			log.Printf("[DEBUG] Deleting GitHub issue label %s/%s/%s", owner, repository, name)

			_, err := client.Issues.DeleteLabel(ctx, owner, repository, o["name"].(string))
			if err != nil {
				return err
			}
		}
	}

	// update
	for name, n := range nMap {
		if o, ok := oMap[name]; ok {
			if o["name"] != n["name"] || o["color"] != n["color"] || o["description"] != n["description"] {
				log.Printf("[DEBUG] Updating GitHub issue label %s/%s/%s", owner, repository, name)

				_, _, err := client.Issues.EditLabel(ctx, owner, repository, name, &github.Label{
					Name:        github.String(n["name"].(string)),
					Color:       github.String(n["color"].(string)),
					Description: github.String(n["description"].(string)),
				})
				if err != nil {
					return err
				}
			}
		}
	}

	log.Printf("[DEBUG] Reading GitHub issue labels for %s/%s", owner, repository)

	labels, err := githubIssueLabels(client, owner, repository)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Found %d GitHub issue labels for %s/%s", len(labels), owner, repository)
	log.Printf("[DEBUG] Labels: %v", labels)

	log.Printf("[DEBUG] Filtering GitHub issue labels for %s/%s", owner, repository)

	filtered := make([]map[string]interface{}, 0)
	for _, label := range labels {
		name := strings.ToLower(label["name"].(string))
		_, oOk := oMap[name]
		_, nOk := nMap[name]
		if oOk || nOk {
			filtered = append(filtered, label)
		}
	}

	log.Printf("[DEBUG] Filtered %d GitHub issue labels for %s/%s", len(filtered), owner, repository)
	log.Printf("[DEBUG] Filtered labels: %v", filtered)

	d.SetId(repository)

	err = d.Set("label", filtered)
	if err != nil {
		return err
	}

	return nil
}

func githubIssueLabels(client *github.Client, owner, repository string) ([]map[string]interface{}, error) {
	ctx := context.WithValue(context.Background(), ctxId, repository)

	options := &github.ListOptions{
		PerPage: maxPerPage,
	}

	labels := make([]map[string]interface{}, 0)

	for {
		ls, resp, err := client.Issues.ListLabels(ctx, owner, repository, options)
		if err != nil {
			return nil, err
		}
		for _, l := range ls {
			label := make(map[string]interface{})

			label["name"] = l.GetName()
			label["color"] = l.GetColor()
			label["description"] = l.GetDescription()
			label["url"] = l.GetURL()

			labels = append(labels, label)
		}

		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	return labels, nil
}
