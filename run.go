package dreck

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"
)

// sanitize checks the run command s to see if a respects our white list.
// It is also check for a maximum length of 64, allow what isRun matches, but disallow ..
func sanitize(s string) bool {
	if len(s) > 64 {
		return false
	}
	ok := isRun(s)
	if !ok {
		return false
	}

	// Extra check for .. because the regexp doesn't catch that.
	if strings.Contains("..", s) {
		return false
	}

	return true
}

func (d Dreck) run(req types.IssueCommentOuter, cmdType, cmdValue string) error {

	// Due to $reasons cmdValue may be prefixed with spaces and a :, strip those off, cmdValue should
	// then start with a slash.
	pos := strings.Index(cmdValue, "/")
	if pos < 0 {
		return fmt.Errorf("illegal run command %s", cmdValue)
	}
	run := cmdValue[pos:]

	log.Infof("%s wants to run %s for issue #%d\n", req.Comment.User.Login, run, req.Issue.Number)

	parts := strings.Fields(run) // simple split
	if len(parts) == 0 {
		return fmt.Errorf("illegal run command %s", run)
	}

	cmd := exec.Command(parts[0], parts[1:]...)

	// Get stdout, errors will go to Caddy log.
	buf, err := cmd.Output()
	if err != nil {
		return err
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	body := fmt.Sprintf("The command %s has run and output the following on its standard output", run)
	body += "~~~\n" + string(buf) + "\n~~~\n"

	comment := githubIssueComment(body)
	client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, comment)

	return nil
}

// isRun checks our whitelist.
var isRun = regexp.MustCompile(`^[-a-zA-Z0-9 ./]+$`).MatchString
