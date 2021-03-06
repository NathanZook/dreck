package types

type Repository struct {
	Owner Owner  `json:"owner"`
	Name  string `json:"name"`
}

type Owner struct {
	Login string `json:"login"`
	Type  string `json:"type"`
}

type PullRequest struct {
	Number int `json:"number"`
}

type InstallationRequest struct {
	Installation ID `json:"installation"`
}

type ID struct {
	ID int `json:"id"`
}

type PullRequestOuter struct {
	Repository  Repository  `json:"repository"`
	PullRequest PullRequest `json:"pull_request"`
	Action      string      `json:"action"`
	InstallationRequest
	Changes map[string]map[string]string `json:"changes"`
}

type IssueCommentOuter struct {
	Repository Repository `json:"repository"`
	Comment    Comment    `json:"comment"`
	Action     string     `json:"action"`
	Issue      Issue      `json:"issue"`
	InstallationRequest
}

type IssueLabel struct {
	Name string `json:"name"`
}

type Issue struct {
	Labels []IssueLabel `json:"labels"`
	Number int          `json:"number"`
	Title  string       `json:"title"`
	Locked bool         `json:"locked"`
	State  string       `json:"state"`
}

type Comment struct {
	Body     string `json:"body"`
	IssueURL string `json:"issue_url"`
	User     struct {
		Login string `json:"login"`
	}
}

type CommentAction struct {
	Type  string
	Value string
}

// DreckConfig holds the configuration from the top-level owners file.
type DreckConfig struct {
	Aliases   []string
	Features  []string
	Reviewers []string
	Approvers []string
}

// PullRequestToIssueComment converts one type to another. This is not a full copy, but copies
// enough elements to make things work from a pull request.
func PullRequestToIssueComment(pr PullRequestOuter) IssueCommentOuter {
	ico := IssueCommentOuter{}
	ico.Repository = pr.Repository
	ico.Issue.Number = pr.PullRequest.Number
	ico.InstallationRequest = pr.InstallationRequest

	return ico
}
