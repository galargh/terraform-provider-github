package github

import (
	"context"

	"github.com/google/go-github/v52/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubIssueLabels() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubIssueLabelsCreate,
		Read:   resourceGithubIssueLabelsRead,
		Update: resourceGithubIssueLabelsUpdate,
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
			"authoritative": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Take authority over all repository labels.",
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
							Required:    true,
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
	repoName := d.Id()
	authoritative := d.Get("authoritative").(bool)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	options := &github.ListOptions{
		PerPage: 100,
	}

	labels := []interface{}{}

	names := []string{}
	if !authoritative {
		for _, l := range d.Get("label").(*schema.Set).List() {
			names = append(names, l.(map[string]interface{})["name"].(string))
		}
	}

	contains := func(s []string, e string) bool {
		for _, a := range s {
			if a == e {
				return true
			}
		}
		return false
	}

	for {
		ls, resp, err := client.Issues.ListLabels(ctx, owner, repoName, options)
		if err != nil {
			return err
		}
		for _, l := range ls {
			if authoritative || contains(names, l.GetName()) {
				labels = append(labels, map[string]interface{}{
					"name":        l.GetName(),
					"color":       l.GetColor(),
					"description": l.GetDescription(),
					"url":         l.GetURL(),
				})
			}
		}

		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	err := d.Set("repository", repoName)
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
	repoName := d.Get("repository").(string)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	labels := d.Get("label").(*schema.Set)

	for _, label := range labels.List() {
		l := label.(map[string]interface{})

		name := l["name"].(string)

		_, err := client.Issues.DeleteLabel(ctx, owner, repoName, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceGithubIssueLabelsUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	o, n := d.GetChange("label")
	oMap := map[string]map[string]interface{}{}
	nMap := map[string]map[string]interface{}{}
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
			_, err := client.Issues.DeleteLabel(ctx, owner, repoName, name)
			if err != nil {
				return err
			}
		}
	}

	// create
	for name, m := range nMap {
		if _, ok := oMap[name]; !ok {
			_, _, err := client.Issues.CreateLabel(ctx, owner, repoName, &github.Label{
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
			_, _, err := client.Issues.EditLabel(ctx, owner, repoName, name, &github.Label{
				Name:        github.String(name),
				Color:       github.String(m["color"].(string)),
				Description: github.String(m["description"].(string)),
			})
			if err != nil {
				return err
			}
		}
	}

	d.SetId(repoName)

	return resourceGithubIssueLabelsRead(d, meta)
}

func resourceGithubIssueLabelsCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	labels := d.Get("label").(*schema.Set)
	for _, label := range labels.List() {
		l := label.(map[string]interface{})
		_, _, err := client.Issues.CreateLabel(ctx, owner, repoName, &github.Label{
			Name:        github.String(l["name"].(string)),
			Color:       github.String(l["color"].(string)),
			Description: github.String(l["description"].(string)),
		})
		if err != nil {
			return err
		}
	}

	d.SetId(repoName)

	return resourceGithubIssueLabelsRead(d, meta)
}
