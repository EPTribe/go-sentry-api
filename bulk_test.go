package sentry

import (
	"fmt"
	"testing"

	"github.com/getsentry/raven-go"
)

func TestBulkResourceModifyDelete(t *testing.T) {
	t.Parallel()
	org, err := client.GetOrganization("sentry")
	if err != nil {
		t.Fatal(err)
	}
	team, err := client.CreateTeam(org, "test team for go project", nil)
	if err != nil {
		t.Fatal(err)
	}
	project, err := client.CreateProject(org, team, "Test python project issues", nil)
	if err != nil {
		t.Fatal(err)
	}
	dsnkey, err := client.CreateClientKey(org, project, "testing key")
	if err != nil {
		t.Fatal(err)
	}

	ravenClient, err := raven.New(dsnkey.DSN.Secret)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i <= 100; i++ {
		ravenClient.CaptureMessage(fmt.Sprintf("Testing message %d", i), nil, nil)
	}

	t.Run("Fetch all messages for project", func(t *testing.T) {
		issues, link, err := client.GetIssues(org, project)
		if err != nil {
			t.Error(err)
		}

		ids := make([]string, 0)
		for _, issue := range issues {
			ids = append(ids, *issue.ID)
		}

		for link.Next.Results {
			for _, issue := range issues {
				ids = append(ids, *issue.ID)
			}
			link, err = client.GetPage(link.Next, &issues)
			if err != nil {
				t.Error(err)
			}
		}

		t.Run("Modify all messages to be resolved", func(t *testing.T) {
			resolved := Resolved
			resp, err := client.BulkMutateIssues(org, project, IssueBulkRequest{
				Status: &resolved,
			}, &ids, nil)

			if err != nil {
				t.Skip(err)
			}
			if resp.Status != nil {
				if *resp.Status != Resolved {
					t.Error("Should have made this resolved")
				}
			}

		})

		t.Run("Delete all of the messages", func(t *testing.T) {
			err := client.BulkDeleteIssues(org, project, ids)
			if err != nil {
				t.Error(err)
			}
		})
	})

	if err := client.DeleteTeam(org, team); err != nil {
		t.Error(err)
	}
}
