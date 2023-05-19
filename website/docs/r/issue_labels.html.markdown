---
layout: "github"
page_title: "GitHub: github_issue_labels"
description: |-
  Provides GitHub issue labels resource.
---

# github_issue_labels

Provides GitHub issue labels resource.

This resource allows you to create and manage issue labels within your
GitHub organization.

~> Note: github_issue_labels cannot be used in conjunction with github_issue_label on the same labels in the authoritative mode or they will fight over what your policy should be.

## Example Usage

```hcl
# Create a new, red colored label
resource "github_issue_labels" "test_repo" {
  repository = "test-repo"

  label {
    name  = "Urgent"
    color = "FF0000"
  }

  label {
    name  = "Critical"
    color = "FF0000"
  }
}
```

## Argument Reference

The following arguments are supported:

* `repository` - (Required) The GitHub repository

* `authoritative` - (Optional) Whether or not this resource is authoritative. If true, this resource will remove any labels that are not specified in the configuration. If false, this resource will only add labels that are specified in the configuration. Defaults to true.

* `name` - (Required) The name of the label.

* `color` - (Required) A 6 character hex code, **without the leading #**, identifying the color of the label.

* `description` - (Optional) A short description of the label.

* `url` - (Computed) The URL to the issue label

## Import

GitHub Issue Labels can be imported using the repository `name`, e.g.

```
$ terraform import github_issue_labels.test_repo test_repo
```
