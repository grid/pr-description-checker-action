package main

import (
	"context"
	"os"
	"regexp"
	"strconv"
	"strings"
	"fmt"

	"github.com/google/go-github/v42/github"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

var (
	markdownCommentRegex = regexp.MustCompile(`<!--[\s\S]*?-->`)
)

type config struct {
	githubToken              string
	templatePath             string
	exemptLabels             []string
	comment                  bool
	commentEmptyDescription  string
	commentTemplateNotFilled string
	commentGithubToken       string

	repoOwner string
	repoName  string
	prNumber  int
}

func generateConfig() *config {
	cfg := config{}

	cfg.githubToken = githubactions.GetInput("repo-token")
	cfg.templatePath = githubactions.GetInput("template-path")
	cfg.exemptLabels = strings.Split(githubactions.GetInput("exempt-labels"), ",")
	cfg.comment, _ = strconv.ParseBool(githubactions.GetInput("comment"))
	cfg.commentEmptyDescription = githubactions.GetInput("comment-empty-description")
	cfg.commentTemplateNotFilled = githubactions.GetInput("comment-template-not-filled")
	cfg.commentGithubToken = githubactions.GetInput("comment-github-token")

	cfg.prNumber, _ = strconv.Atoi(githubactions.GetInput("pr-number"))
	cfg.repoOwner = githubactions.GetInput("repo-owner")
	cfg.repoName = githubactions.GetInput("repo-name")

	return &cfg
}

func fetchTemplate() (string, error) {
	data, err := os.ReadFile(cfg.templatePath)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func splitByHeaders(markdown string) []map[string]string {
	// Regular expression to match Markdown headers
	re := regexp.MustCompile(`(?m)^(#{1,6}) (.+)$`)

	// Find all header matches
	matches := re.FindAllStringSubmatchIndex(markdown, -1)

	// Split the Markdown into sections based on headers
	var sections []map[string]string
	for i, match := range matches {
		headerLevel := len(markdown[match[2]:match[3]]) // Number of `#`
		headerText := markdown[match[4]:match[5]]       // Header text

		// Determine the start of the content (skip the header line)
		start := match[1] + 1 // Start right after the header line
		var end int
		if i+1 < len(matches) {
			end = matches[i+1][0]
		} else {
			end = len(markdown)
		}
		var content string
		if start > end {
			content = ""
		} else {
			content = strings.TrimSpace(markdown[start:end]) // Trim spaces from content
		}
		sections = append(sections, map[string]string{
			"header_level": fmt.Sprintf("%d", headerLevel),
			"header_text":  headerText,
			"content":      content,
		})
	}

	return sections
}

func newGithubClient(token string) *github.Client {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func normalizeDescription(description string) string {
	desc := strings.Replace(description, "\r\n", "\n", -1)
	desc = markdownCommentRegex.ReplaceAllString(desc, "")
	desc = strings.TrimSpace(desc)

	return desc
}

var cfg *config

func main() {
	cfg = generateConfig()

	template, err := fetchTemplate()
	template = normalizeDescription(template)

	if err != nil {
		githubactions.Infof("Failed to fetch template: %s, will continue without template", err)
	}

	githubClient := newGithubClient(cfg.githubToken)

	pr, _, _ := githubClient.PullRequests.Get(context.Background(), cfg.repoOwner, cfg.repoName, cfg.prNumber)

	skipCheck := false
	for _, label := range pr.Labels {
		for _, exemptLabel := range cfg.exemptLabels {
			if label.GetName() == strings.Trim(exemptLabel, " ") {
				skipCheck = true
				break
			}
		}
	}

	if skipCheck {
		githubactions.Infof("Skipping check because of exempt label")
		os.Exit(0)
	}

	description := normalizeDescription(pr.GetBody())
	var errorMsg string
	if len(description) == 0 {
		errorMsg = cfg.commentEmptyDescription
	} else if len(description) <= len(template) {
		errorMsg = cfg.commentTemplateNotFilled
	}

	if errorMsg != "" {
		if cfg.comment {
			if cfg.commentGithubToken != "" {
				githubClient = newGithubClient(cfg.commentGithubToken)
			}

			_, _, err := githubClient.Issues.CreateComment(context.Background(), cfg.repoOwner, cfg.repoName, cfg.prNumber, &github.IssueComment{
				Body: &errorMsg,
			})

			if err != nil {
				githubactions.Fatalf("Failed to create comment: %s", err)
			}
		}

		githubactions.Fatalf(errorMsg)
	}

	githubactions.Infof("Description is valid")
}
