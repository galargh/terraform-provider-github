package github

import (
	"context"

	"github.com/google/go-github/v52/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubIssueLabels() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubIssueLabelsCreateOrUpdate,
		Read:   resourceGithubIssueLabelsRead,
		Update: resourceGithubIssueLabelsCreateOrUpdate,
		Delete: resourceGithubIssueLabelsDelete,
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
	repository := d.Get("repository").(string)

	labels, err := githubIssueLabels(client, owner, repository)
	if err != nil {
		return err
	}

	err = d.Set("label", labels)
	if err != nil {
		return err
	}

	return nil
}

func resourceGithubIssueLabelsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repository := d.Get("repository").(string)
	ctx := context.WithValue(context.Background(), ctxId, repository)

	labels := d.Get("label").(*schema.Set).List()

	for _, label := range labels {
		l := label.(map[string]interface{})

		name := l["name"].(string)

		_, err := client.Issues.DeleteLabel(ctx, owner, repository, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceGithubIssueLabelsCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repository := d.Get("repository").(string)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	o, n := d.GetChange("label")
	oMap := make(map[string]map[string]interface{})
	nMap := make(map[string]map[string]interface{})
	for _, raw := range o.(*schema.Set).List() {
		m := raw.(map[string]interface{})
		name := m["name"].(string)
		oMap[name] = m
	}
	for _, raw := range n.(*schema.Set).List() {
		m := raw.(map[string]interface{})
		name := m["name"].(string)
		nMap[name] = m
	}

	// delete
	for name := range oMap {
		if _, ok := nMap[name]; !ok {
			_, err := client.Issues.DeleteLabel(ctx, owner, repository, name)
			if err != nil {
				return err
			}
		}
	}

	// create
	for name, m := range nMap {
		if _, ok := oMap[name]; !ok {
			_, _, err := client.Issues.CreateLabel(ctx, owner, repository, &github.Label{
				Name:        github.String(name),
				Color:       github.String(m["color"].(string)),
				Description: github.String(m["description"].(string)),
			})
			if err != nil {
				return err
			}
		}
	}

	// update
	for name, m := range nMap {
		if _, ok := oMap[name]; ok {
			_, _, err := client.Issues.EditLabel(ctx, owner, repository, name, &github.Label{
				Name:        github.String(name),
				Color:       github.String(m["color"].(string)),
				Description: github.String(m["description"].(string)),
			})
			if err != nil {
				return err
			}
		}
	}

	labels, err := githubIssueLabels(client, owner, repository)
	if err != nil {
		return err
	}

	filtered := make([]map[string]interface{}, 0)
	for _, label := range labels {
		name := label["name"].(string)
		_, oOk := oMap[name]
		_, nOk := nMap[name]
		if oOk || nOk {
			filtered = append(filtered, label)
		}
	}

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
