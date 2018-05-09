package dreck

import (
	"context"
	"fmt"
	"strings"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/github"
)

// handlePullRequestDCO handles the DCO check. I.e. a PR must have commits for Signed-off-by.
func (d Dreck) handlePullRequestDCO(req types.PullRequestOuter) error {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	hasUnsignedCommits, err := hasUnsigned(req, client)
	if err != nil {
		return err
	}

	if hasUnsignedCommits {
		issue, _, labelErr := client.Issues.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)

		if labelErr != nil {
			return labelErr
		}
		log.Infof("Current labels: %s", issue.Labels)

		if hasNoDcoLabel(issue) == false {
			log.Info("Applying label")
			_, _, assignLabelErr := client.Issues.AddLabelsToIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, []string{"no-dco"})
			if assignLabelErr != nil {
				return assignLabelErr
			}

			link := fmt.Sprintf("https://github.com/%s/%s/blob/master/CONTRIBUTING.md", req.Repository.Owner.Login, req.Repository.Name)
			body := `Thank you for your contribution. I've just checked and your commit doesn't appear to be signed-off.
That's something we need before your Pull Request can be merged. Please see our [contributing guide](` + link + `).`

			comment := githubIssueComment(body)

			comment, resp, err := client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, comment)
			if err != nil {
				return err
			}
			log.Infof("%s %s", comment, resp.Rate)
		}
	} else {
		issue, _, labelErr := client.Issues.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)

		if labelErr != nil {
			return labelErr
		}

		if hasNoDcoLabel(issue) {
			log.Info("Removing label")
			_, removeLabelErr := client.Issues.RemoveLabelForIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, "no-dco")
			if removeLabelErr != nil {
				return removeLabelErr
			}
		}
	}
	return nil
}

func hasNoDcoLabel(issue *github.Issue) bool {
	if issue != nil {
		for _, label := range issue.Labels {
			if label.GetName() == "no-dco" {
				return true
			}
		}
	}
	return false
}

func hasUnsigned(req types.PullRequestOuter, client *github.Client) (bool, error) {
	hasUnsigned := false
	ctx := context.Background()

	listOpts := &github.ListOptions{Page: 0}

	commits, resp, err := client.PullRequests.ListCommits(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, listOpts)
	if err != nil {
		return hasUnsigned, fmt.Errorf("getting PR %d\n%s", req.PullRequest.Number, err.Error())
	}

	log.Warningf("Rate limiting: %s", resp.Rate)

	for _, commit := range commits {
		if commit.Commit != nil && commit.Commit.Message != nil {
			if isSigned(*commit.Commit.Message) == false {
				hasUnsigned = true
			}
		}
	}

	return hasUnsigned, nil
}

func isSigned(msg string) bool {
	return strings.Contains(msg, "Signed-off-by:")
}

// handlePullRequestReview will look at the (first 5) files of a PR, retrieve the nearest OWNERS files
// merge all the reviewers and randomly pick a reviewer that should be assigned for this PR.
func (d Dreck) handlePullRequestReviewers(req types.PullRequestOuter) error {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	listOpts := &github.ListOptions{Page: 0}
	files, resp, err := client.PullRequests.ListFiles(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, listOpts)
	if err != nil {
		return fmt.Errorf("getting PR %d\n%s", req.PullRequest.Number, err.Error())
	}

	pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)
	if err != nil {
		return fmt.Errorf("getting PR %d\n%s", req.PullRequest.Number, err.Error())
	}

	log.Warningf("Rate limiting: %s", resp.Rate)

	victims := make(map[string]bool) // possible reviewers

File:
	for _, f := range files {
		paths := ownersPaths(*f.Filename, d.owners)
		// Find nearest OWNERS files.
		for _, p := range paths {
			buf, err := githubFile(req.Repository.Owner.Login, req.Repository.Name, p)
			if err != nil {
				continue
			}

			var config types.DreckConfig
			if err := parseConfig(buf, &config); err != nil {
				continue
			}
			for _, r := range config.Reviewers {
				victims[r] = true
			}
			continue File
		}
	}

	log.Infof("Looking for reviewers in %v, excluding %s", victims, *pull.User.Login)
	// This randomizes for us, pick first non PR author.
	victim := ""
	for v, _ := range victims {
		if v != *pull.User.Login {
			victim = v
			break
		}
	}

	if victim != "" {
		rev := github.ReviewersRequest{Reviewers: []string{victim}}
		if _, _, err := client.PullRequests.RequestReviewers(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, rev); err != nil {
			// set comment?
			return err
		}
	}

	body := "Thank you for your contribution. I've just checked the *%s* files to find a suitable reviewer."
	if victim != "" {
		body += " This search was succesful and I've asked **%s** for a review."
		body = fmt.Sprintf(body, d.owners, victim)
	} else {
		body += " Alas, this search was *not* succesful."
		body = fmt.Sprintf(body, d.owners)
	}

	comment := githubIssueComment(body)
	comment, resp, err = client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, comment)
	log.Infof("%s", resp.Rate)
	return err
}
